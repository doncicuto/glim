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

	resty "github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"

	"github.com/Songmu/prompter"
	"github.com/spf13/cobra"
)

// NewUserCmd - TODO comment
var userPasswdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change a Glim user account password",
	Run: func(cmd *cobra.Command, args []string) {
		passwdBody := models.JSONPasswdBody{}

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/users/%d/passwd", url, userID)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		if !AmIManager(token) && uint32(WhichIsMyTokenUID(token)) != userID {
			fmt.Println("Only users with manager role can change other users passwords")
		}

		if cmd.Flags().Changed("password") {
			fmt.Println("WARNING! Using --password via the CLI is insecure.")
		}

		if uint32(WhichIsMyTokenUID(token)) == userID {
			oldPassword := prompter.Password("Old password")
			if oldPassword == "" {
				fmt.Println("Error password required")
				os.Exit(1)
			}
			passwdBody.OldPassword = oldPassword
		}

		// Prompt for password if needed
		if !cmd.Flags().Changed("password") {
			if !passwordStdin {
				password = prompter.Password("New password")
				if password == "" {
					fmt.Println("Error password required")
					os.Exit(1)
				}
				confirmPassword := prompter.Password("Confirm password")
				if password != confirmPassword {
					fmt.Println("Error passwords don't match")
					os.Exit(1)
				}
			}
		}

		passwdBody.Password = password

		// Rest API authentication
		client := resty.New()
		client.SetAuthToken(token.AccessToken)
		// TODO - We should verify server's certificate
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(passwdBody).
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

		fmt.Println("Password changed")
	},
}

func init() {
	userPasswdCmd.Flags().Uint32VarP(&userID, "uid", "i", 0, "User account id")
	userPasswdCmd.Flags().StringVarP(&password, "password", "p", "", "New user password")
	// Mark required flags
	cobra.MarkFlagRequired(userPasswdCmd.Flags(), "uid")
}
