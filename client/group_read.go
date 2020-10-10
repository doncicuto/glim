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
