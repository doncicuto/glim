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
	"fmt"
	"os"

	"github.com/doncicuto/glim/models"
	resty "github.com/go-resty/resty/v2"

	"github.com/Songmu/prompter"
	"github.com/spf13/cobra"
)

// NewUserCmd - TODO comment
var userPasswdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change a Glim user account password",
	Run: func(cmd *cobra.Command, args []string) {
		passwdBody := models.JSONPasswdBody{}

		// Glim server URL
		url := os.Getenv("GLIM_URI")
		if url == "" {
			url = serverAddress
		}

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
		client.SetRootCertificate(tlscacert)

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
