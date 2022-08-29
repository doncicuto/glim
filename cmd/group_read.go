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

package cmd

import (
	"fmt"
	"os"

	"github.com/doncicuto/glim/models"
	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getGIDFromGroupName(group string, url string) uint {
	token := ReadCredentials()
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	client := RestClient(token.AccessToken)
	endpoint := fmt.Sprintf("%s/v1/groups/%s/gid", url, group)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(models.GroupID{}).
		SetError(&types.APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
		os.Exit(1)
	}

	result := resp.Result().(*models.GroupID)
	return uint(result.ID)
}

func getGroup(id uint) {
	// Glim server URL
	url := viper.GetString("server")

	endpoint := fmt.Sprintf("%s/v1/groups/%d", url, id)
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
		SetResult(models.Group{}).
		SetError(&types.APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
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
	// Glim server URL
	url := viper.GetString("server")

	// Read credentials
	token := ReadCredentials()
	endpoint := fmt.Sprintf("%s/v1/groups", url)
	// Check expiration
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	// Rest API authentication
	client := RestClient(token.AccessToken)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult([]models.Group{}).
		SetError(&types.APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
		os.Exit(1)
	}

	fmt.Printf("%-6s %-20s %-35s %-50s\n",
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
		fmt.Printf("%-6d %-20s %-35s %-50s\n",
			result.ID,
			truncate(*result.Name, 20),
			truncate(*result.Description, 35),
			truncate(members, 50))
	}
}

func GetGroupInfo() {
	gid := viper.GetUint("gid")
	group := viper.GetString("group")
	if gid != 0 {
		getGroup(gid)
		os.Exit(0)
	}
	if group != "" {
		url := viper.GetString("server")
		gid = getGIDFromGroupName(group, url)
		getGroup(gid)
		os.Exit(0)
	}
	getGroups()
	os.Exit(0)
}

// ListGroupCmd - TODO comment
var listGroupCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Glim groups",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		GetGroupInfo()
	},
}

func init() {
	listGroupCmd.Flags().UintP("gid", "i", 0, "group id")
	listGroupCmd.Flags().StringP("group", "g", "", "group name")
}
