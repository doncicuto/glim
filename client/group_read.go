//
// Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>
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

	"github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"
	"github.com/spf13/cobra"
)

func getGroup(id int) {
	endpoint := fmt.Sprintf("%s/groups/%d", url, id)
	// Read credentials
	token := ReadCredentials()

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
		SetResult(models.Group{}).
		SetError(&APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
		os.Exit(1)
	}

	result := resp.Result().(*models.Group)
	fmt.Printf("%-15s %-100s\n", "Group:", *result.Name)
	fmt.Printf("%-15s %-100d\n", " GID:", result.ID)
	fmt.Printf("%-15s %-100s\n\n", " Description:", *result.Description)

	fmt.Printf("%-15s\n", "Members:")
	fmt.Printf("====\n")
	for _, member := range result.Members {
		fmt.Printf("%-15s %-100d\n", " UID:", member.ID)
		fmt.Printf("%-15s %-100s\n", " Username:", *member.Username)
		fmt.Printf("----\n")
	}
}

func getGroups() {
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
		SetResult([]models.Group{}).
		SetError(&APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
		os.Exit(1)
	}

	fmt.Printf("%-10s %-20s %-35s %-50s\n",
		"GID",
		"GROUP",
		"DESCRIPTION",
		"MEMBERS",
	)

	results := resp.Result().(*[]models.Group)
	for _, result := range *results {
		members := "none"
		if len(result.Members) > 0 {
			members = ""
			for i, member := range result.Members {
				if i == len(result.Members)-1 {
					members += *member.Username
				} else {
					members += *member.Username + ", "
				}
			}
		}
		fmt.Printf("%-10d %-20s %-35s %-50s\n",
			result.ID,
			truncate(*result.Name, 20),
			truncate(*result.Description, 35),
			truncate(members, 50))
	}
}

// ListGroupCmd - TODO comment
var listGroupCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Glim groups",
	Run: func(cmd *cobra.Command, args []string) {
		getGroups()
	},
}
