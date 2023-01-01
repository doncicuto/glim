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
	"strings"

	"github.com/doncicuto/glim/common"
	"github.com/doncicuto/glim/models"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getUIDFromUsername(client *resty.Client, username string, url string) (uint, error) {
	endpoint := fmt.Sprintf("%s/v1/users/%s/uid", url, username)
	resp, err := client.R().
		SetHeader(contentTypeHeader, appJson).
		SetResult(models.UserID{}).
		SetError(&common.APIError{}).
		Get(endpoint)

	if err != nil {
		return 0, fmt.Errorf(common.CantConnectMessage, err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
	}

	result := resp.Result().(*models.UserID)
	return uint(result.ID), nil
}

func getUser(cmd *cobra.Command, id uint, jsonOutput bool) error {
	// Glim server URL
	url := viper.GetString("server")
	endpoint := fmt.Sprintf("%s/v1/users/%d", url, id)

	// Get credentials
	token, err := GetCredentials(url)
	if err != nil {
		return err
	}

	// Rest API authentication
	client := RestClient(token.AccessToken)

	resp, err := client.R().
		SetHeader(contentTypeHeader, appJson).
		SetResult(models.UserInfo{}).
		SetError(&common.APIError{}).
		Get(endpoint)

	if err != nil {
		return fmt.Errorf(common.CantConnectMessage, err)
	}

	if resp.IsError() {
		return fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
	}

	// memberOf := "none"
	result := resp.Result().(*models.UserInfo)

	if jsonOutput {
		encodeUserToJson(cmd, result)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "\n%-15s %-100d\n", "UID:", result.ID)
		fmt.Fprint(cmd.OutOrStdout(), sectionSeparator)
		fmt.Fprintf(cmd.OutOrStdout(), descrFormat, "Username:", result.Username)
		fmt.Fprintf(cmd.OutOrStdout(), descrFormat, "Name:", strings.Join([]string{result.GivenName, result.Surname}, " "))
		fmt.Fprintf(cmd.OutOrStdout(), descrFormat, "Email:", result.Email)
		fmt.Fprintf(cmd.OutOrStdout(), accountFormat, "Manager:", result.Manager)
		fmt.Fprintf(cmd.OutOrStdout(), accountFormat, "Read-Only:", result.Readonly)
		fmt.Fprintf(cmd.OutOrStdout(), accountFormat, "Locked:", result.Locked)
		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %s\n", "SSH Public Key:", result.SSHPublicKey)
		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %s\n", "JPEG Photo:", truncate(result.JPEGPhoto, 100))
		fmt.Fprintf(cmd.OutOrStdout(), "----\n")
		if len(result.MemberOf) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Member of: \n")
			for _, group := range result.MemberOf {
				fmt.Fprintf(cmd.OutOrStdout(), " * GID: %-4d Name: %-100s\n", group.ID, group.Name)
			}
		}
	}
	return nil
}

func getUsers(cmd *cobra.Command, jsonOutput bool) error {
	// Glim server URL
	url := viper.GetString("server")
	endpoint := fmt.Sprintf("%s/v1/users", url)

	// Get credentials
	token, err := GetCredentials(url)
	if err != nil {
		return err
	}

	// Rest API authentication
	client := RestClient(token.AccessToken)

	resp, err := client.R().
		SetHeader(contentTypeHeader, appJson).
		SetResult([]models.UserInfo{}).
		SetError(&common.APIError{}).
		Get(endpoint)

	if err != nil {
		return fmt.Errorf(common.CantConnectMessage, err)
	}

	if resp.IsError() {
		return fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
	}

	results := resp.Result().(*[]models.UserInfo)

	if jsonOutput {
		encodeUsersToJson(cmd, results)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%-6s %-15s %-20s %-20s %-20s %-8s %-8s %-8s\n",
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

			fmt.Fprintf(cmd.OutOrStdout(), "%-6d %-15s %-20s %-20s %-20s %-8v %-8v %-8v\n",
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
	return nil
}

func GetUserInfo(cmd *cobra.Command) error {
	url := viper.GetString("server")
	uid := viper.GetUint("uid")
	username := viper.GetString("username")
	jsonOutput := viper.GetBool("json")
	if uid != 0 {
		err := getUser(cmd, uid, jsonOutput)
		if err != nil {
			return err
		}
		return nil
	}
	if username != "" {
		// Get credentials
		token, err := GetCredentials(url)
		if err != nil {
			return err
		}
		client := RestClient(token.AccessToken)
		uid, err = getUIDFromUsername(client, username, url)
		if err != nil {
			return err
		}
		err = getUser(cmd, uid, jsonOutput)
		if err != nil {
			return err
		}
		return nil
	}

	// Get credentials
	token, err := GetCredentials(url)
	if err != nil {
		return err
	}

	if AmIPlainUser(token) {
		tokenUID, err := WhichIsMyTokenUID(token)
		if err != nil {
			return err
		}

		uid = tokenUID
		err = getUser(cmd, uid, jsonOutput)
		if err != nil {
			return err
		}
		return nil
	}

	err = getUsers(cmd, jsonOutput)
	if err != nil {
		return err
	}
	return nil
}

// ListUserCmd is a Cobra Command that lists users available in Glim
func ListUserCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List Glim user accounts",
		PreRun: func(cmd *cobra.Command, _ []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return GetUserInfo(cmd)
		},
	}

	cmd.Flags().UintP("uid", "i", 0, "user account id")
	cmd.Flags().StringP("username", "u", "", "username")
	addGlimPersistentFlags(cmd)

	return cmd
}
