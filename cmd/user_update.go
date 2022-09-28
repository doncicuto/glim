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

// NewUserCmd - TODO comment
var updateUserCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Glim user account",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {

		var trueValue = true
		var falseValue = false

		// Read credentials and check expiration
		token := ReadCredentials()
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Get uid and username
		uid := viper.GetUint("uid")
		username := viper.GetString("username")

		// JSON output?
		jsonOutput := viper.GetBool("json")

		if uid == 0 && username == "" {
			error := "you must specify either the user account id or a username"
			printError(error, jsonOutput)
			os.Exit(1)
		}

		// Glim server URL
		url := viper.GetString("server")

		if uid == 0 && username != "" {
			uid = getUIDFromUsername(username, url, jsonOutput)
		}

		// Validate email
		email := viper.GetString("email")
		if email != "" {
			if _, err := mail.ParseAddress(email); err != nil {
				error := "email should have a valid format"
				printError(error, jsonOutput)
				os.Exit(1)
			}
		}

		// Check if both manager and readonly have been set
		manager := viper.GetBool("manager")
		readonly := viper.GetBool("readonly")
		if manager && readonly {
			error := "a Glim account cannot be both manager and readonly at the same time"
			printError(error, jsonOutput)
			os.Exit(1)
		}

		// Check if both remove and replace flags have been set
		replace := viper.GetBool("replace")
		remove := viper.GetBool("remove")
		if replace && remove {
			error := "replace and remove flags are mutually exclusive"
			printError(error, jsonOutput)
			os.Exit(1)
		}

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

		userBody := models.JSONUserBody{
			Username:     username,
			Name:         strings.Join([]string{viper.GetString("firstname"), viper.GetString("lastname")}, " "),
			GivenName:    viper.GetString("firstname"),
			Surname:      viper.GetString("lastname"),
			Email:        viper.GetString("email"),
			SSHPublicKey: viper.GetString("ssh-public-key"),
			MemberOf:     viper.GetString("groups"),
			JPEGPhoto:    jpegPhoto,
		}

		if viper.GetBool("manager") {
			userBody.Manager = &trueValue
			userBody.Readonly = &falseValue
		}

		if viper.GetBool("readonly") {
			userBody.Manager = &falseValue
			userBody.Readonly = &trueValue
		}

		if viper.GetBool("lock") {
			userBody.Locked = &trueValue
		}

		if viper.GetBool("unlock") {
			userBody.Locked = &falseValue
		}

		if viper.GetBool("plainuser") {
			userBody.Manager = &falseValue
			userBody.Readonly = &falseValue
		}

		if replace {
			userBody.ReplaceMembersOf = true
		}

		if remove {
			userBody.RemoveMembersOf = true
		}

		// Rest API authentication
		client := RestClient(token.AccessToken)
		endpoint := fmt.Sprintf("%s/v1/users/%d", url, uid)
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(userBody).
			SetError(&types.APIError{}).
			Put(endpoint)

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

		printMessage("User updated", jsonOutput)
	},
}

func init() {
	updateUserCmd.Flags().StringP("username", "u", "", "username")
	updateUserCmd.Flags().StringP("firstname", "f", "", "first name")
	updateUserCmd.Flags().StringP("lastname", "l", "", "last name")
	updateUserCmd.Flags().StringP("email", "e", "", "email")
	updateUserCmd.Flags().StringP("ssh-public-key", "k", "", "SSH Public Key")
	updateUserCmd.Flags().StringP("jpeg-photo", "j", "", "path to avatar file (jpg, png)")
	updateUserCmd.Flags().StringP("groups", "g", "", "comma-separated list of group names. ")
	updateUserCmd.Flags().Bool("manager", false, "Glim manager account?")
	updateUserCmd.Flags().Bool("readonly", false, "Glim readonly account?")
	updateUserCmd.Flags().Bool("plainuser", false, "Glim plain user account. User can read and modify its own user account information but not its group membership.")
	updateUserCmd.Flags().Bool("replace", false, "replace groups with those specified with -g. Groups are appended to those that the user is a member of by default")
	updateUserCmd.Flags().Bool("remove", false, "remove group membership with those specified with -g.")
	updateUserCmd.Flags().Bool("lock", false, "lock account (cannot log in)")
	updateUserCmd.Flags().Bool("unlock", false, "unlock account (can log in)")
	updateUserCmd.Flags().UintP("uid", "i", 0, "user account id")
}
