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
var newGroupCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Glim group",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		// json output?
		jsonOutput := viper.GetBool("json")

		// Glim server URL
		url := viper.GetString("server")
		endpoint := fmt.Sprintf("%s/v1/groups", url)

		// Get credentials
		token, err := GetCredentials(url)
		if err != nil {
			printError(err.Error(), jsonOutput)
			os.Exit(1)
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(models.JSONGroupBody{
				Name:        viper.GetString("group"),
				Description: viper.GetString("description"),
				Members:     viper.GetString("members"),
			}).
			SetError(&types.APIError{}).
			Post(endpoint)

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

		printMessage("Group created", jsonOutput)
	},
}

func init() {
	newGroupCmd.Flags().StringP("group", "g", "", "our group name")
	newGroupCmd.Flags().StringP("description", "d", "", "our group description")
	newGroupCmd.Flags().StringP("members", "m", "", "comma-separated list of usernames e.g: admin,tux")
	newGroupCmd.MarkFlagRequired("group")
}
