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
	resty "github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

// newGroupCmd - TODO comment
var newGroupCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Glim group",
	Run: func(cmd *cobra.Command, args []string) {

		// Glim server URL
		url := os.Getenv("GLIM_URI")
		if url == "" {
			url = serverAddress
		}

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/groups", url)
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
			SetBody(models.JSONGroupBody{
				Name:        groupName,
				Description: groupDesc,
				Members:     groupMembers,
			}).
			SetError(&APIError{}).
			Post(endpoint)

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
			os.Exit(1)
		}

		fmt.Println("Group created")
	},
}

func init() {
	newGroupCmd.Flags().StringVarP(&groupName, "name", "n", "", "our group name")
	newGroupCmd.Flags().StringVarP(&groupDesc, "description", "d", "", "our group description")
	newGroupCmd.Flags().StringVarP(&groupMembers, "members", "m", "", "comma-separated list of usernames e.g: admin,tux")

	// Mark required flags
	cobra.MarkFlagRequired(newGroupCmd.Flags(), "name")
}
