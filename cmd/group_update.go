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

	"github.com/doncicuto/glim/models"
	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newGroupCmd - TODO comment
var updateGroupCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Glim group",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		// Glim server URL
		url := viper.GetString("server")

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/v1/groups/%d", url, viper.GetUint("gid"))
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(models.JSONGroupBody{
				Name:           viper.GetString("name"),
				Description:    viper.GetString("description"),
				Members:        viper.GetString("members"),
				ReplaceMembers: viper.GetBool("replace"),
			}).
			SetError(&types.APIError{}).
			Put(endpoint)

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			os.Exit(1)
		}

		fmt.Println("Group updated")
	},
}

func init() {
	updateGroupCmd.Flags().UintP("gid", "g", 0, "group id")
	updateGroupCmd.Flags().StringP("name", "n", "", "our group name")
	updateGroupCmd.Flags().StringP("description", "d", "", "our group description")
	updateGroupCmd.Flags().StringP("members", "m", "", "comma-separated list of usernames e.g: admin,tux")
	updateGroupCmd.Flags().Bool("replace", false, "Replace group members with those specified with -m. Usernames are appended to members by default")
	updateGroupCmd.MarkFlagRequired("gid")
}
