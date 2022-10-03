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

	"github.com/Songmu/prompter"
	"github.com/doncicuto/glim/types"
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
		url := viper.GetString("server")

		confirm := prompter.YesNo("Do you really want to delete this group?", false)
		if !confirm {
			os.Exit(1)
		}

		// json output?
		jsonOutput := viper.GetBool("json")

		// Get credentials
		token, err := GetCredentials(url)
		if err != nil {
			printError(err.Error(), jsonOutput)
			os.Exit(1)
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		gid := viper.GetUint("gid")
		group := viper.GetString("group")
		if gid == 0 && group != "" {
			gid, err = getGIDFromGroupName(client, group, url)
			if err != nil {
				printError(err.Error(), jsonOutput)
				os.Exit(1)
			}
		}
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
			error := fmt.Sprintf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			printError(error, jsonOutput)
			os.Exit(1)
		}

		printMessage("Group deleted", jsonOutput)
	},
}

func init() {
	deleteGroupCmd.Flags().UintP("gid", "i", 0, "group id")
	deleteGroupCmd.Flags().StringP("group", "g", "", "group name")
}
