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
