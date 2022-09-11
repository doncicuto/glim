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
	"net/mail"
	"os"
	"strings"

	"github.com/doncicuto/glim/models"
	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// csvCreateUsersCmd - TODO comment
var csvCreateUsersCmd = &cobra.Command{
	Use:   "create",
	Short: "Create users from a CSV file",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {

		// Read and open file
		users := readUsersFromCSV()

		if len(users) == 0 {
			fmt.Println("no users where found in CSV file")
			os.Exit(1)
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

		fmt.Println("")

		for _, user := range users {
			username := *user.Username
			// Validate email
			email := *user.Email
			if email != "" {
				if _, err := mail.ParseAddress(email); err != nil {
					fmt.Printf("%s: skipped, email should have a valid format\n", username)
					continue
				}
			}
			// Check if both manager and readonly has been set
			manager := *user.Manager
			readonly := *user.Readonly
			if manager && readonly {
				fmt.Printf("%s: skipped, cannot be both manager and readonly at the same time\n", username)
				continue
			}

			password := *user.Password
			locked := *user.Locked || password == ""

			// JpegPhoto
			jpegPhoto := ""
			jpegPhotoPath := *user.JPEGPhoto
			if jpegPhotoPath != "" {
				photo, err := JPEGToBase64(jpegPhotoPath)
				if err != nil {
					fmt.Printf("%s: skipped, could not convert JPEG photo to Base64 %v\n", username, err)
					continue
				}
				jpegPhoto = *photo
			}

			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(models.JSONUserBody{
					Username:     username,
					Password:     password,
					Name:         strings.Join([]string{*user.GivenName, *user.Surname}, " "),
					GivenName:    *user.GivenName,
					Surname:      *user.Surname,
					Email:        *user.Email,
					SSHPublicKey: *user.SSHPublicKey,
					MemberOf:     *user.Groups,
					JPEGPhoto:    jpegPhoto,
					Manager:      &manager,
					Readonly:     &readonly,
					Locked:       &locked,
				}).
				SetError(&types.APIError{}).
				Post(endpoint)

			if err != nil {
				fmt.Printf("Error connecting with Glim: %v\n", err)
				os.Exit(1)
			}

			if resp.IsError() {
				fmt.Printf("%s: skipped, %v\n", username, resp.Error().(*types.APIError).Message)
				continue
			}
			fmt.Printf("%s: successfully created\n", username)
		}

		fmt.Printf("\nCreate from CSV finished!\n")
	},
}

func init() {
	csvCreateUsersCmd.Flags().StringP("file", "f", "", "path to CSV file, use README to know more about the format")
	csvCreateUsersCmd.MarkFlagRequired("file")
}
