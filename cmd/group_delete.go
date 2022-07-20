/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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

package client

import (
	"fmt"
	"os"

	"github.com/Songmu/prompter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DeleteGroupCmd - TODO comment
var deleteGroupCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a Glim group",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {

		confirm := prompter.YesNo("Do you really want to delete this group?", false)
		if !confirm {
			os.Exit(1)
		}

		// Glim server URL
		url := viper.GetString("server")

		// Read credentials
		gid := viper.GetUint("gid")
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/v1/groups/%d", url, gid)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetError(&APIError{}).
			Delete(endpoint)

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
			os.Exit(1)
		}

		fmt.Println("Group deleted")
	},
}

func init() {
	deleteGroupCmd.Flags().UintP("gid", "g", 0, "group id")
	deleteGroupCmd.MarkFlagRequired("gid")
}
