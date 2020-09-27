//
// Copyright © 2020 Muultipla Devops <devops@muultipla.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package client

import (
	"crypto/tls"
	"fmt"
	"os"

	resty "github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"
	"github.com/spf13/cobra"
)

var (
	groupName, groupDesc, groupMembers string
)

// newGroupCmd - TODO comment
var newGroupCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Glim group",
	Run: func(cmd *cobra.Command, args []string) {

		url := "https://127.0.0.1:1323" // TODO - This should not be hardcoded

		// Glim server URL
		if len(args) > 0 {
			url = args[0]
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
		// TODO - We should verify server's certificate
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(models.NewGroup{
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
	rootCmd.AddCommand(newGroupCmd)

	newGroupCmd.Flags().StringVarP(&groupName, "name", "n", "", "our group name")
	newGroupCmd.Flags().StringVarP(&groupDesc, "description", "d", "", "our group description")
	newGroupCmd.Flags().StringVarP(&groupMembers, "members", "m", "", "comma-separated list of usernames e.g: manager,tux")

	// Mark required flags
	cobra.MarkFlagRequired(newGroupCmd.Flags(), "name")
}
