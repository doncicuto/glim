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

// DeleteUserCmd - TODO comment
var deleteUserCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a Glim user account",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {

		confirm := prompter.YesNo("Do you really want to delete this user?", false)
		if !confirm {
			os.Exit(1)
		}

		url := viper.GetString("server")
		uid := viper.GetUint("uid")

		// Read credentials and check if token needs refresh
		token := ReadCredentials()
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)
		endpoint := fmt.Sprintf("%s/v1/users/%d", url, uid)
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

		fmt.Println("User account deleted")
	},
}

func init() {
	deleteUserCmd.Flags().UintP("uid", "i", 0, "user account id")
	deleteUserCmd.MarkPersistentFlagRequired("uid")
}
