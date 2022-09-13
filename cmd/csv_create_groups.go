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

// csvCreateGroupsCmd - TODO comment
var csvCreateGroupsCmd = &cobra.Command{
	Use:   "create",
	Short: "Create groups from a CSV file",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {

		// Read and open file
		groups := readGroupsFromCSV()

		if len(groups) == 0 {
			fmt.Println("no groups where found in CSV file")
			os.Exit(1)
		}

		// Glim server URL
		url := viper.GetString("server")

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/v1/groups", url)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		fmt.Println("")

		for _, group := range groups {
			name := *group.Name
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(models.JSONGroupBody{
					Name:        *group.Name,
					Description: *group.Description,
					Members:     *group.GroupMembers,
				}).
				SetError(&types.APIError{}).
				Post(endpoint)

			if err != nil {
				fmt.Printf("Error connecting with Glim: %v\n", err)
				os.Exit(1)
			}

			if resp.IsError() {
				fmt.Printf("%s: skipped, %v\n", name, resp.Error().(*types.APIError).Message)
				continue
			}
			fmt.Printf("%s: successfully created\n", name)
		}

		fmt.Printf("\nCreate from CSV finished!\n")
	},
}

func init() {
	csvCreateGroupsCmd.Flags().StringP("file", "f", "", "path to CSV file, use README to know more about the format")
	csvCreateGroupsCmd.MarkFlagRequired("file")
}
