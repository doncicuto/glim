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

// csvDeleteUsersCmd - TODO comment
var csvDeleteUsersCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove users included in a CSV file",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		// json output?
		jsonOutput := viper.GetBool("json")

		// Read and open file
		users := readUsersFromCSV(jsonOutput)

		messages := []string{}

		if len(users) == 0 {
			error := "no users where found in CSV file"
			printError(error, jsonOutput)
			os.Exit(1)
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

		for _, user := range users {
			username := *user.Username
			uid := user.ID

			if username == "" && uid <= 0 {
				error := fmt.Sprintf("UID %d: skipped, invalid username and uid\n", uid)
				messages = append(messages, error)
				continue
			}

			if username != "" {
				endpoint := fmt.Sprintf("%s/v1/users/%s/uid", url, username)
				resp, err := client.R().
					SetHeader("Content-Type", "application/json").
					SetResult(models.User{}).
					SetError(&types.APIError{}).
					Get(endpoint)

				if err != nil {
					error := fmt.Sprintf("Error connecting with Glim: %v\n", err)
					printError(error, jsonOutput)
					os.Exit(1)
				}

				if resp.IsError() {
					error := fmt.Sprintf("%s: skipped, %v\n", username, resp.Error().(*types.APIError).Message)
					messages = append(messages, error)
					continue
				}

				result := resp.Result().(*models.User)

				if result.ID != uid && uid != 0 {
					error := fmt.Sprintf("%s: skipped, username and uid found in CSV doesn't match\n", username)
					messages = append(messages, error)
					continue
				}
				uid = result.ID
			}

			// Delete using API
			endpoint := fmt.Sprintf("%s/v1/users/%d", url, uid)
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetError(&types.APIError{}).
				Delete(endpoint)

			if err != nil {
				fmt.Printf("Error connecting with Glim: %v\n", err)
				os.Exit(1)
			}

			if resp.IsError() {
				if username != "" {
					error := fmt.Sprintf("%s: skipped, %v\n", username, resp.Error().(*types.APIError).Message)
					messages = append(messages, error)
				} else {
					error := fmt.Sprintf("UID %d: skipped, %v\n", uid, resp.Error().(*types.APIError).Message)
					messages = append(messages, error)
				}
				continue
			}

			message := fmt.Sprintf("%s: successfully removed\n", username)
			messages = append(messages, message)

		}

		printCSVMessages(messages, jsonOutput)
		if !jsonOutput {
			fmt.Printf("\nRemove from CSV finished!\n")
		}
	},
}

func init() {
	csvDeleteUsersCmd.Flags().StringP("file", "f", "", "path to CSV file, use README to know more about the format")
	csvDeleteUsersCmd.MarkFlagRequired("file")
}
