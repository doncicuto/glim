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

	"github.com/doncicuto/glim/models"
	"github.com/doncicuto/glim/types"

	"github.com/Songmu/prompter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUserCmd - TODO comment
var userPasswdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change a Glim user account password",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		passwdBody := models.JSONPasswdBody{}

		url := viper.GetString("server")
		uid := viper.GetUint("uid")
		username := viper.GetString("username")

		// Check expiration
		token := ReadCredentials()
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// JSON output?
		jsonOutput := viper.GetBool("json")

		if uid == 0 {
			if username != "" {
				uid = getUIDFromUsername(username, url, jsonOutput)
			} else {
				uid = uint(WhichIsMyTokenUID(token))
			}
		}

		if !AmIManager(token) && uint(WhichIsMyTokenUID(token)) != uid {
			error := "Only users with manager role can change other users passwords"
			printError(error, jsonOutput)
			os.Exit(1)
		}

		if uint(WhichIsMyTokenUID(token)) == uid {
			oldPassword := prompter.Password("Old password")
			if oldPassword == "" {
				error := "Error password required"
				printError(error, jsonOutput)
				os.Exit(1)
			}
			passwdBody.OldPassword = oldPassword
		}

		password := viper.GetString("password")

		if password != "" {
			fmt.Println("WARNING! Using --password via the CLI is insecure.")
		} else {
			passwordStdin := viper.GetBool("passwd-stdin")
			if !passwordStdin {
				password = prompter.Password("New password")
				if password == "" {
					error := "Error password required"
					printError(error, jsonOutput)
					os.Exit(1)
				}
				confirmPassword := prompter.Password("Confirm password")
				if password != confirmPassword {
					error := "Error passwords don't match"
					printError(error, jsonOutput)
					os.Exit(1)
				}
			}
		}

		passwdBody.Password = password

		// Rest API authentication
		client := RestClient(token.AccessToken)
		endpoint := fmt.Sprintf("%s/v1/users/%d/passwd", url, uid)
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(passwdBody).
			SetError(&types.APIError{}).
			Post(endpoint)

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			os.Exit(1)
		}

		fmt.Println("Password changed")
	},
}

func init() {
	userPasswdCmd.Flags().UintP("uid", "i", 0, "User account id")
	userPasswdCmd.Flags().StringP("username", "u", "", "username")
	userPasswdCmd.Flags().StringP("password", "p", "", "New user password")
	userPasswdCmd.Flags().Bool("password-stdin", false, "Take the password from stdin")
}
