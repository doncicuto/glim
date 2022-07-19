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
	"strings"

	"github.com/doncicuto/glim/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getUser(id uint) {
	// Glim server URL
	url := viper.GetString("server")
	endpoint := fmt.Sprintf("%s/users/%d", url, id)
	// Read credentials
	token := ReadCredentials()

	// Check expiration
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	// Rest API authentication
	client := RestClient(token.AccessToken)

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

	// memberOf := "none"
	result := resp.Result().(*models.UserInfo)

	fmt.Printf("\n%-15s %-100d\n", "UID:", result.ID)
	fmt.Println("====")
	fmt.Printf("%-15s %-100s\n", "Username:", result.Username)
	fmt.Printf("%-15s %-100s\n", "Name:", strings.Join([]string{result.GivenName, result.Surname}, " "))
	fmt.Printf("%-15s %-100s\n", "Email:", result.Email)
	fmt.Printf("%-15s %-8v\n", "Manager:", result.Manager)
	fmt.Printf("%-15s %-8v\n", "Read-Only:", result.Readonly)
	fmt.Printf("%-15s %-8v\n", "Locked:", result.Locked)
	fmt.Println("----")
	if len(result.MemberOf) > 0 {
		fmt.Println("Member of: ")
		for _, group := range result.MemberOf {
			fmt.Printf(" * GID: %-4d Name: %-100s\n", group.ID, group.Name)
		}
	}
}

func getUsers() {
	// Glim server URL
	url := viper.GetString("server")

	// Read credentials
	token := ReadCredentials()
	endpoint := fmt.Sprintf("%s/users", url)
	// Check expiration
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	// Rest API authentication
	client := RestClient(token.AccessToken)

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

	fmt.Printf("%-6s %-15s %-20s %-20s %-20s %-8s %-8s %-8s\n",
		"UID",
		"USERNAME",
		"FULLNAME",
		"EMAIL",
		"GROUPS",
		"MANAGER",
		"READONLY",
		"LOCKED",
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

		fmt.Printf("%-6d %-15s %-20s %-20s %-20s %-8v %-8v %-8v\n",
			result.ID,
			truncate(result.Username, 15),
			truncate(strings.Join([]string{result.GivenName, result.Surname}, " "), 20),
			truncate(result.Email, 20),
			truncate(memberOf, 20),
			result.Manager,
			result.Readonly,
			result.Locked,
		)
	}
}

// ListUserCmd - TODO comment
var listUserCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Glim user accounts",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		tlscacert := viper.GetString("tlscacert")
		_, err := os.Stat(tlscacert)
		if os.IsNotExist(err) {
			fmt.Println("Could not find required CA pem file to validate authority")
			os.Exit(1)
		}

		uid := viper.GetUint("uid")
		if uid != 0 {
			getUser(uid)
			os.Exit(0)
		}
		getUsers()
	},
}

func init() {
	listUserCmd.Flags().UintP("uid", "i", 0, "user account id")
}