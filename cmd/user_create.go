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
	"bufio"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/doncicuto/glim/models"
	"github.com/doncicuto/glim/types"
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
		// json output?
		jsonOutput := viper.GetBool("json")

		// Validate email
		email := viper.GetString("email")
		if email != "" {
			if _, err := mail.ParseAddress(email); err != nil {
				error := "email should have a valid format"
				printError(error, jsonOutput)
				os.Exit(1)
			}
		}

		// Check if both manager and readonly has been set

		manager := viper.GetBool("manager")
		readonly := viper.GetBool("readonly")

		if manager && readonly {
			error := "a Glim account cannot be both manager and readonly at the same time"
			printError(error, jsonOutput)
			os.Exit(1)
		}

		plainuser := viper.GetBool("plainuser")
		if plainuser {
			manager = false
			readonly = false
		}

		// Prompt for password if needed
		password := viper.GetString("password")
		passwordStdin := viper.GetBool("password-stdin")
		locked := viper.GetBool("lock")

		if password == "" && !passwordStdin && !locked {
			password = prompter.Password("Password")
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
		} else {
			switch {
			case password != "" && !passwordStdin:
				fmt.Println("WARNING! Using --password via the CLI is insecure. Use --password-stdin.")

			case password != "" && passwordStdin:
				error := "--password and --password-stdin are mutually exclusive"
				printError(error, jsonOutput)
				os.Exit(1)

			case passwordStdin:
				// Reference: https://flaviocopes.com/go-shell-pipes/
				info, err := os.Stdin.Stat()
				if err != nil {
					error := "Error reading from stdin"
					printError(error, jsonOutput)
					os.Exit(1)
				}

				if info.Mode()&os.ModeCharDevice != 0 {
					error := "Error expecting password from stdin using a pipe"
					printError(error, jsonOutput)
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
				if password == "" {
					locked = true
				}
			}
		}

		// Glim server URL
		url := viper.GetString("server")

		// Read credentials
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/v1/users", url)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)

		// JpegPhoto
		jpegPhoto := ""
		jpegPhotoPath := viper.GetString("jpeg-photo")
		if jpegPhotoPath != "" {
			photo, err := JPEGToBase64(jpegPhotoPath)
			if err != nil {
				error := fmt.Sprintf("could not convert JPEG photo to Base64 - %v\n", err)
				printError(error, jsonOutput)
				os.Exit(1)
			}
			jpegPhoto = *photo
		}

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(models.JSONUserBody{
				Username:     viper.GetString("username"),
				Password:     password,
				Name:         strings.Join([]string{viper.GetString("firstname"), viper.GetString("lastname")}, " "),
				GivenName:    viper.GetString("firstname"),
				Surname:      viper.GetString("lastname"),
				Email:        viper.GetString("email"),
				SSHPublicKey: viper.GetString("ssh-public-key"),
				MemberOf:     viper.GetString("groups"),
				JPEGPhoto:    jpegPhoto,
				Manager:      &manager,
				Readonly:     &readonly,
				Locked:       &locked,
			}).
			SetError(&types.APIError{}).
			Post(endpoint)

		if err != nil {
			error := fmt.Sprintf("Error connecting with Glim: %v\n", err)
			printError(error, jsonOutput)
			os.Exit(1)
		}

		if resp.IsError() {
			error := fmt.Sprintf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			printError(error, jsonOutput)
			os.Exit(1)
		}

		printMessage("User created", jsonOutput)
	},
}

func init() {
	// newUserCmd.Flags().UintP("uid", "i", 0, "User account id")
	newUserCmd.Flags().StringP("username", "u", "", "username")
	newUserCmd.Flags().StringP("firstname", "f", "", "first name")
	newUserCmd.Flags().StringP("lastname", "l", "", "last name")
	newUserCmd.Flags().StringP("email", "e", "", "email")
	newUserCmd.Flags().StringP("password", "p", "", "password")
	newUserCmd.Flags().StringP("ssh-public-key", "k", "", "SSH Public Key")
	newUserCmd.Flags().StringP("jpeg-photo", "j", "", "path to avatar file (jpg, png)")
	newUserCmd.Flags().StringP("groups", "g", "", "comma-separated list of groups that we want the new user account to be a member of")
	newUserCmd.Flags().Bool("password-stdin", false, "take the password from stdin")
	newUserCmd.Flags().Bool("manager", false, "Glim manager account?")
	newUserCmd.Flags().Bool("readonly", false, "Glim readonly account?")
	newUserCmd.Flags().Bool("plainuser", false, "Glim plain user account. User can read and modify its own user account information but not its group membership.")
	newUserCmd.Flags().Bool("lock", false, "lock account (no password will be set, user cannot log in)")
	newUserCmd.Flags().Bool("unlock", false, "unlock account (can log in)")
	newUserCmd.MarkFlagRequired("username")
}
