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

	"github.com/badoux/checkmail"
	resty "github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"
	"github.com/spf13/cobra"
)

// NewUserCmd - TODO comment
var updateUserCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Glim user account",
	Run: func(cmd *cobra.Command, args []string) {

		var trueValue = true
		var falseValue = false

		// Validate email
		if email != "" {
			if err := checkmail.ValidateFormat(email); err != nil {
				fmt.Println("email should have a valid format")
				os.Exit(1)
			}
		}

		// Check if both manager and readonly has been set
		if manager && readonly {
			fmt.Println("a Glim account cannot be both manager and readonly at the same time")
			os.Exit(1)
		}

		// Glim server URL
		if len(args) > 0 {
			url = args[0]
		}

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/users/%d", url, userID)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		userBody := models.JSONUserBody{
			Username: username,
			Fullname: fullname,
			Email:    email,
			MemberOf: groups,
		}

		if manager {
			userBody.Manager = &trueValue
			userBody.Readonly = &falseValue
		}

		if readonly {
			userBody.Manager = &falseValue
			userBody.Readonly = &trueValue
		}

		if plainuser {
			userBody.Manager = &falseValue
			userBody.Readonly = &falseValue
		}

		if replaceMembersOf {
			userBody.ReplaceMembersOf = true
		}

		// Rest API authentication
		client := resty.New()
		client.SetAuthToken(token.AccessToken)
		// TODO - We should verify server's certificate
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(userBody).
			SetError(&APIError{}).
			Put(endpoint)

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
			os.Exit(1)
		}

		fmt.Println("User updated")
	},
}

func init() {
	updateUserCmd.Flags().Uint32VarP(&userID, "uid", "i", 0, "User account id")
	updateUserCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	updateUserCmd.Flags().StringVarP(&fullname, "fullname", "f", "", "Fullname")
	updateUserCmd.Flags().StringVarP(&email, "email", "e", "", "Email")
	updateUserCmd.Flags().StringVarP(&groups, "groups", "g", "", "Comma-separated list of new groups that we want the user account to be a member of. ")
	updateUserCmd.Flags().BoolVar(&manager, "manager", false, "Glim manager account?")
	updateUserCmd.Flags().BoolVar(&readonly, "readonly", false, "Glim readonly account?")
	updateUserCmd.Flags().BoolVar(&plainuser, "plainuser", false, "Glim plain user account. User can read and modify its own user account information but not its group membership.")
	updateUserCmd.Flags().BoolVar(&replaceMembersOf, "replace", false, "Replace groups with those specified with -g. Groups are appended to those that the user is a member of by default")

	// Mark required flags
	cobra.MarkFlagRequired(updateUserCmd.Flags(), "uid")
}
