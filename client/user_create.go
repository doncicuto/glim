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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/badoux/checkmail"
	resty "github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"
	"github.com/spf13/cobra"
)

// NewUserCmd - TODO comment
var newUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Glim user account",
	Run: func(cmd *cobra.Command, args []string) {

		// Validate email
		if err := checkmail.ValidateFormat(email); err != nil {
			fmt.Println("email should have a valid format")
			os.Exit(1)
		}

		// Check if both manager and readonly has been set
		if manager && readonly {
			fmt.Println("a Glim account cannot be both manager and readonly at the same time")
			os.Exit(1)
		}

		// Prompt for password if needed
		if !cmd.Flags().Changed("password") {
			if !passwordStdin {
				password = prompter.Password("Password")
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
		} else {
			fmt.Println("WARNING! Using --password via the CLI is insecure. Use --password-stdin.")
		}

		// Check if password has been sent in stdin
		if passwordStdin {
			if password != "" {
				fmt.Println("--password and --password-stdin are mutually exclusive")
				os.Exit(1)
			} else {
				// Reference: https://flaviocopes.com/go-shell-pipes/
				info, err := os.Stdin.Stat()
				if err != nil {
					fmt.Println("Error reading from stdin")
					os.Exit(1)
				}

				if info.Mode()&os.ModeCharDevice != 0 {
					fmt.Println("Error expecting password from stdin using a pipe")
					os.Exit(1)
				}

				reader := bufio.NewReader(os.Stdin)
				var output []rune

				for {
					input, _, err := reader.ReadRune()
					if err != nil && err == io.EOF {
						break
					}
					output = append(output, input)
				}

				password = strings.TrimSuffix(string(output), "\n")
			}
		}

		// Glim server URL
		if len(args) > 0 {
			url = args[0]
		}

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
		client.SetRootCertificate(tlscacert)

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(models.JSONUserBody{
				Username: username,
				Password: password,
				Fullname: fullname,
				Email:    email,
				MemberOf: groups,
				Manager:  &manager,
				Readonly: &readonly,
			}).
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

		fmt.Println("User created")
	},
}

func init() {
	newUserCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	newUserCmd.Flags().StringVarP(&fullname, "fullname", "f", "", "Fullname")
	newUserCmd.Flags().StringVarP(&email, "email", "e", "", "Email")
	newUserCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	newUserCmd.Flags().StringVarP(&groups, "groups", "g", "", "Comma-separated list of groups that we want the new user account to be a member of")
	newUserCmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Take the password from stdin")
	newUserCmd.Flags().BoolVar(&manager, "manager", false, "Glim manager account?")
	newUserCmd.Flags().BoolVar(&readonly, "readonly", false, "Glim readonly account?")

	// Mark required flags
	cobra.MarkFlagRequired(newUserCmd.Flags(), "username")
	cobra.MarkFlagRequired(newUserCmd.Flags(), "fullname")
	cobra.MarkFlagRequired(newUserCmd.Flags(), "email")
}
