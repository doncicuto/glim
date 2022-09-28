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
	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// csvDeleteGroupsCmd - TODO comment
var csvDeleteGroupsCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove groups included in a CSV file",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		// json output?
		jsonOutput := viper.GetBool("json")
		messages := []string{}

		// Read and open file
		groups := readGroupsFromCSV(jsonOutput)

		if len(groups) == 0 {
			error := "no groups where found in CSV file"
			printError(error, jsonOutput)
			os.Exit(1)
		}

		// Glim server URL
		url := viper.GetString("server")

		// Read credentials
		token := ReadCredentials()
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		for _, group := range groups {
			name := *group.Name
			gid := group.ID

			if name == "" && gid <= 0 {
				error := fmt.Sprintf("GID %d: skipped, invalid group name and gid\n", gid)
				messages = append(messages, error)
				continue
			}

			if name != "" {
				endpoint := fmt.Sprintf("%s/v1/groups/%s/gid", url, name)
				resp, err := client.R().
					SetHeader("Content-Type", "application/json").
					SetResult(models.Group{}).
					SetError(&types.APIError{}).
					Get(endpoint)

				if err != nil {
					error := fmt.Sprintf("Error connecting with Glim: %v\n", err)
					printError(error, jsonOutput)
					os.Exit(1)
				}

				if resp.IsError() {
					error := fmt.Sprintf("%s: skipped, %v\n", name, resp.Error().(*types.APIError).Message)
					messages = append(messages, error)
					continue
				}

				result := resp.Result().(*models.Group)

				if result.ID != gid && gid != 0 {
					error := fmt.Sprintf("%s: skipped, group name and gid found in CSV doesn't match\n", name)
					messages = append(messages, error)
					continue
				}
				gid = result.ID
			}

			// Delete using API
			endpoint := fmt.Sprintf("%s/v1/groups/%d", url, gid)
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetError(&types.APIError{}).
				Delete(endpoint)

			if err != nil {
				error := fmt.Sprintf("Error connecting with Glim: %v\n", err)
				printError(error, jsonOutput)
				os.Exit(1)
			}

			if resp.IsError() {
				if name != "" {
					error := fmt.Sprintf("%s: skipped, %v\n", name, resp.Error().(*types.APIError).Message)
					messages = append(messages, error)
				} else {
					error := fmt.Sprintf("GID %d: skipped, %v\n", gid, resp.Error().(*types.APIError).Message)
					messages = append(messages, error)
				}
				continue
			}
			message := fmt.Sprintf("%s: successfully removed\n", name)
			messages = append(messages, message)
		}

		printCSVMessages(messages, jsonOutput)
		if !jsonOutput {
			fmt.Printf("\nRemove from CSV finished!\n")
		}

	},
}

func init() {
	csvDeleteGroupsCmd.Flags().StringP("file", "f", "", "path to CSV file, use README to know more about the format")
	csvDeleteGroupsCmd.MarkFlagRequired("file")
}
