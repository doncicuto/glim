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

	"github.com/doncicuto/glim/common"
	"github.com/doncicuto/glim/models"

	"github.com/Songmu/prompter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUserCmd - TODO comment
func UserPasswdCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "passwd",
		Short: "Change a Glim user account password",
		RunE: func(cmd *cobra.Command, _ []string) error {
			passwdBody := models.JSONPasswdBody{}

			url := viper.GetString("server")
			uid := viper.GetUint("uid")
			username := viper.GetString("username")

			// Get credentials
			token, err := GetCredentials(url)
			if err != nil {
				return err
			}

			// JSON output?
			jsonOutput := viper.GetBool("json")

			tokenUID, err := WhichIsMyTokenUID(token)
			if err != nil {
				return err
			}

			client := RestClient(token.AccessToken)
			if uid == 0 {
				if username != "" {
					uid, err = getUIDFromUsername(client, username, url)
					if err != nil {
						return fmt.Errorf("only users with manager role can change other users passwords")
					}
				} else {
					uid = tokenUID
				}
			}

			// fmt.Println(!AmIManager(token), tokenUID, uid)
			if !AmIManager(token) && tokenUID != uid {
				return fmt.Errorf("only users with manager role can change other users passwords")
			}

			if tokenUID == uid {
				oldPassword := prompter.Password("Old password")
				if oldPassword == "" {
					return fmt.Errorf("password required")
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
						return fmt.Errorf("password required")
					}
					confirmPassword := prompter.Password("Confirm password")
					if password != confirmPassword {
						return fmt.Errorf("passwords don't match")
					}
				}
			}

			passwdBody.Password = password

			endpoint := fmt.Sprintf("%s/v1/users/%d/passwd", url, uid)
			resp, clientError := client.R().
				SetHeader(contentTypeHeader, appJson).
				SetBody(passwdBody).
				SetError(&common.APIError{}).
				Post(endpoint)

			if err := checkAPICallResponse(clientError, resp); err != nil {
				return err
			}

			printCmdMessage(cmd, "Password changed", jsonOutput)
			return nil
		},
	}

	// General flags
	cmd.Flags().UintP("uid", "i", 0, "User account id")
	cmd.Flags().StringP("username", "u", "", "username")
	cmd.Flags().StringP("password", "p", "", "New user password")

	// Boolean flags
	cmd.Flags().Bool("password-stdin", false, "Take the password from stdin")
	addGlimPersistentFlags(cmd)
	viper.BindPFlags(cmd.Flags())

	return cmd
}
