/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@sologitops.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/doncicuto/glim/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/doncicuto/glim/common"
)

func CsvDeleteGroupsCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove groups included in a CSV file",
		PreRun: func(cmd *cobra.Command, _ []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// json output?
			jsonOutput := viper.GetBool("json")
			messages := []string{}

			// Read and open file
			groups, err := readGroupsFromCSV(jsonOutput, "gid, name")
			if err != nil {
				return err
			}

			// Glim server URL
			url := viper.GetString("server")

			// Get credentials
			token, err := GetCredentials(url)
			if err != nil {
				printError(err.Error(), jsonOutput)
				os.Exit(1)
			}

			// Rest API authentication
			client := RestClient(token.AccessToken)

			for _, group := range groups {
				name := *group.Name
				gid := group.ID

				if name == "" && gid <= 0 {
					messages = append(messages, fmt.Sprintf("GID %d: skipped, invalid group name and gid\n", gid))
					continue
				}

				if name != "" {
					endpoint := fmt.Sprintf("%s/v1/groups/%s/gid", url, name)
					resp, err := client.R().
						SetHeader(contentTypeHeader, appJson).
						SetResult(models.Group{}).
						SetError(&common.APIError{}).
						Get(endpoint)

					if err != nil {
						return fmt.Errorf(common.CantConnectMessage, err)
					}

					if resp.IsError() {
						messages = append(messages, fmt.Sprintf("%s: skipped, %v\n", name, resp.Error().(*common.APIError).Message))
						continue
					}

					result := resp.Result().(*models.Group)

					if result.ID != gid && gid != 0 {
						messages = append(messages, fmt.Sprintf("%s: skipped, group name and gid found in CSV doesn't match\n", name))
						continue
					}
					gid = result.ID
				}

				// Delete using API
				endpoint := fmt.Sprintf("%s/v1/groups/%d", url, gid)
				resp, err := client.R().
					SetHeader(contentTypeHeader, appJson).
					SetError(&common.APIError{}).
					Delete(endpoint)

				if err != nil {
					return fmt.Errorf(common.CantConnectMessage, err)
				}

				if resp.IsError() {
					if name != "" {
						messages = append(messages, fmt.Sprintf("%s: skipped, %v", name, resp.Error().(*common.APIError).Message))
					} else {
						messages = append(messages, fmt.Sprintf("GID %d: skipped, %v", gid, resp.Error().(*common.APIError).Message))
					}
					continue
				}
				message := fmt.Sprintf("%s: successfully removed\n", name)
				messages = append(messages, message)
			}

			printCSVMessages(cmd, messages, jsonOutput)
			if !jsonOutput {
				printCmdMessage(cmd, "Remove from CSV finished!", jsonOutput)
			}

			return nil
		},
	}

	addGlimPersistentFlags(cmd)
	cmd.Flags().StringP("file", "f", "", "path to CSV file, use README to know more about the format")
	cmd.MarkFlagRequired("file")
	return cmd
}
