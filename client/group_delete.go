/*
Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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
	resty "github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

// DeleteGroupCmd - TODO comment
var deleteGroupCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a Glim group",
	Run: func(cmd *cobra.Command, args []string) {

		confirm := prompter.YesNo("Do you really want to delete this group?", false)
		if !confirm {
			os.Exit(1)
		}

		// Glim server URL
		url := os.Getenv("GLIM_URI")
		if url == "" {
			url = serverAddress
		}

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/groups/%d", url, groupID)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := resty.New()
		client.SetAuthToken(token.AccessToken)
		client.SetRootCertificate(tlscacert)

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
	deleteGroupCmd.Flags().Uint32VarP(&groupID, "gid", "i", 0, "group id")

	// Mark required flags
	cobra.MarkFlagRequired(deleteGroupCmd.Flags(), "gid")
}
