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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/badoux/checkmail"
	"github.com/doncicuto/glim/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUserCmd - TODO comment
var newUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Glim user account",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		// Validate email
		email := viper.GetString("email")
		if email != "" {
			if err := checkmail.ValidateFormat(email); err != nil {
				fmt.Println("email should have a valid format")
				os.Exit(1)
			}
		}

		// Check if both manager and readonly has been set

		manager := viper.GetBool("manager")
		readonly := viper.GetBool("readonly")

		if manager && readonly {
			fmt.Println("a Glim account cannot be both manager and readonly at the same time")
			os.Exit(1)
		}

		// Prompt for password if needed
		password := viper.GetString("password")
		passwordStdin := viper.GetBool("password-stdin")
		if password != "" {
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

		locked := len(password) == 0 || viper.GetBool("lock") || !viper.GetBool("unlock")
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(models.JSONUserBody{
				Username:  viper.GetString("username"),
				Password:  password,
				GivenName: viper.GetString("firstname"),
				Surname:   viper.GetString("surname"),
				Email:     viper.GetString("email"),
				MemberOf:  viper.GetString("groups"),
				Manager:   &manager,
				Readonly:  &readonly,
				Locked:    &locked,
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
	// newUserCmd.Flags().UintP("uid", "i", 0, "User account id")
	newUserCmd.Flags().StringP("username", "u", "", "Username")
	newUserCmd.Flags().StringP("firstname", "f", "", "First name")
	newUserCmd.Flags().StringP("lastname", "l", "", "Last name")
	newUserCmd.Flags().StringP("email", "e", "", "Email")
	newUserCmd.Flags().StringP("password", "p", "", "Password")
	newUserCmd.Flags().StringP("groups", "g", "", "Comma-separated list of groups that we want the new user account to be a member of")
	newUserCmd.Flags().Bool("password-stdin", false, "Take the password from stdin")
	newUserCmd.Flags().Bool("manager", false, "Glim manager account?")
	newUserCmd.Flags().Bool("readonly", false, "Glim readonly account?")
	newUserCmd.Flags().Bool("lock", false, "Lock account (cannot log in)")
	newUserCmd.Flags().Bool("unlock", false, "Unlock account (can log in)")
	newUserCmd.MarkFlagRequired("username")
}
