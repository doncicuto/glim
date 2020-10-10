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
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"
	"github.com/spf13/cobra"
)

func getUser(id uint32) {
	endpoint := fmt.Sprintf("%s/users/%d", url, id)
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
		SetResult(models.UserInfo{}).
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

	fmt.Printf("%-6s %-25s %-25s %-25s %-25s %-8s %-8s\n",
		"UID",
		"USERNAME",
		"FULLNAME",
		"EMAIL",
		"GROUPS",
		"MANAGER",
		"READONLY",
	)

	memberOf := "none"
	groups := []string{}

	result := resp.Result().(*models.UserInfo)
	for _, group := range result.MemberOf {
		groups = append(groups, group.Name)
	}

	if len(groups) > 0 {
		memberOf = strings.Join(groups, ",")
	}

	fmt.Printf("%-6d %-25s %-25s %-25s %-25s %-8v %-8v\n",
		result.ID,
		truncate(result.Username, 25),
		truncate(result.Fullname, 25),
		truncate(result.Email, 25),
		truncate(memberOf, 25),
		result.Manager,
		result.Readonly,
	)

}

func getUsers() {
	// Read credentials
	token := ReadCredentials()
	endpoint := fmt.Sprintf("%s/users", url)
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
		SetResult([]models.UserInfo{}).
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

	fmt.Printf("%-6s %-25s %-25s %-25s %-25s %-8s %-8s\n",
		"UID",
		"USERNAME",
		"FULLNAME",
		"EMAIL",
		"GROUPS",
		"MANAGER",
		"READONLY",
	)

	results := resp.Result().(*[]models.UserInfo)

	for _, result := range *results {
		memberOf := "none"
		groups := []string{}

		for _, group := range result.MemberOf {
			groups = append(groups, group.Name)
		}

		if len(groups) > 0 {
			memberOf = strings.Join(groups, ",")
		}

		fmt.Printf("%-6d %-25s %-25s %-25s %-25s %-8v %-8v\n",
			result.ID,
			truncate(result.Username, 25),
			truncate(result.Fullname, 25),
			truncate(result.Email, 25),
			truncate(memberOf, 25),
			result.Manager,
			result.Readonly,
		)
	}
}

// ListUserCmd - TODO comment
var listUserCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Glim user accounts",
	Run: func(cmd *cobra.Command, args []string) {
		getUsers()
	},
}
