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

// newGroupCmd - TODO comment
var updateGroupCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Glim group",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		// json output?
		jsonOutput := viper.GetBool("json")

		// Glim server URL
		url := viper.GetString("server")
		gid := viper.GetUint("gid")
		group := viper.GetString("group")

		if gid == 0 && group == "" {
			error := "you must specify either the group id or name"
			printError(error, jsonOutput)
			os.Exit(1)
		}

		if gid == 0 && group != "" {
			gid = getGIDFromGroupName(group, url, jsonOutput)
		}

		// Read credentials
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
			SetBody(models.JSONGroupBody{
				Name:           viper.GetString("group"),
				Description:    viper.GetString("description"),
				Members:        viper.GetString("members"),
				ReplaceMembers: viper.GetBool("replace"),
			}).
			SetError(&types.APIError{}).
			Put(endpoint)

		if err != nil {
			error := fmt.Sprintf("Error connecting with Glim: %v\n", err)
			printError(error, jsonOutput)
			os.Exit(1)
		}

		if resp.IsError() {
			error := fmt.Sprintf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			printError(error, jsonOutput)
			os.Exit(1)
		}

		printMessage("Group updated", jsonOutput)
	},
}

func init() {
	updateGroupCmd.Flags().UintP("gid", "i", 0, "group id")
	updateGroupCmd.Flags().StringP("group", "g", "", "our group name")
	updateGroupCmd.Flags().StringP("description", "d", "", "our group description")
	updateGroupCmd.Flags().StringP("members", "m", "", "comma-separated list of usernames e.g: admin,tux")
	updateGroupCmd.Flags().Bool("replace", false, "Replace group members with those specified with -m. Usernames are appended to members by default")
}
