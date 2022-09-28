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
	"strings"

	"github.com/doncicuto/glim/models"
	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getUIDFromUsername(username string, url string, jsonOutput bool) uint {
	token := ReadCredentials()
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	client := RestClient(token.AccessToken)
	endpoint := fmt.Sprintf("%s/v1/users/%s/uid", url, username)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(models.UserID{}).
		SetError(&types.APIError{}).
		Get(endpoint)

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

	result := resp.Result().(*models.UserID)
	return uint(result.ID)
}

func getUser(id uint, jsonOutput bool) {
	// Glim server URL
	url := viper.GetString("server")
	endpoint := fmt.Sprintf("%s/v1/users/%d", url, id)
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
		SetError(&types.APIError{}).
		Get(endpoint)

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

	// memberOf := "none"
	result := resp.Result().(*models.UserInfo)

	if jsonOutput {
		encodeUserToJson(result)
	} else {
		fmt.Printf("\n%-15s %-100d\n", "UID:", result.ID)
		fmt.Println("====")
		fmt.Printf("%-15s %-100s\n", "Username:", result.Username)
		fmt.Printf("%-15s %-100s\n", "Name:", strings.Join([]string{result.GivenName, result.Surname}, " "))
		fmt.Printf("%-15s %-100s\n", "Email:", result.Email)
		fmt.Printf("%-15s %-8v\n", "Manager:", result.Manager)
		fmt.Printf("%-15s %-8v\n", "Read-Only:", result.Readonly)
		fmt.Printf("%-15s %-8v\n", "Locked:", result.Locked)
		fmt.Printf("%-15s %s\n", "SSH Public Key:", result.SSHPublicKey)
		fmt.Printf("%-15s %s\n", "JPEG Photo:", truncate(result.JPEGPhoto, 100))
		fmt.Println("----")
		if len(result.MemberOf) > 0 {
			fmt.Println("Member of: ")
			for _, group := range result.MemberOf {
				fmt.Printf(" * GID: %-4d Name: %-100s\n", group.ID, group.Name)
			}
		}
	}
}

func getUsers(jsonOutput bool) {
	// Glim server URL
	url := viper.GetString("server")

	// Read credentials
	token := ReadCredentials()
	endpoint := fmt.Sprintf("%s/v1/users", url)
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
		SetError(&types.APIError{}).
		Get(endpoint)

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

	results := resp.Result().(*[]models.UserInfo)

	if jsonOutput {
		encodeUsersToJson(results)
	} else {
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

}

func GetUserInfo() {
	uid := viper.GetUint("uid")
	username := viper.GetString("username")
	jsonOutput := viper.GetBool("json")
	if uid != 0 {
		getUser(uid, jsonOutput)
		os.Exit(0)
	}
	if username != "" {
		url := viper.GetString("server")
		uid = getUIDFromUsername(username, url, jsonOutput)
		getUser(uid, jsonOutput)
		os.Exit(0)
	}

	// Check expiration
	token := ReadCredentials()
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}
	if AmIPlainUser(token) {
		uid = uint(WhichIsMyTokenUID(token))
		getUser(uid, jsonOutput)
		os.Exit(0)
	}

	getUsers(jsonOutput)
}

// ListUserCmd - TODO comment
var listUserCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Glim user accounts",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		GetUserInfo()
	},
}

func init() {
	listUserCmd.Flags().UintP("uid", "i", 0, "user account id")
	listUserCmd.Flags().StringP("username", "u", "", "username")
}
