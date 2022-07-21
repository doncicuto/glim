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
	"fmt"
	"os"

	"github.com/badoux/checkmail"
	"github.com/doncicuto/glim/models"
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

		// Validate email
		email := viper.GetString("email")
		if email != "" {
			if err := checkmail.ValidateFormat(email); err != nil {
				fmt.Println("email should have a valid format")
				os.Exit(1)
			}
		}

		// Check if both manager and readonly have been set
		manager := viper.GetBool("manager")
		readonly := viper.GetBool("readonly")
		if manager && readonly {
			fmt.Println("a Glim account cannot be both manager and readonly at the same time")
			os.Exit(1)
		}

		// Check if both remove and replace flags have been set
		replace := viper.GetBool("replace")
		remove := viper.GetBool("remove")
		if replace && remove {
			fmt.Println("replace and remove flags are mutually exclusive")
			os.Exit(1)
		}

		// Glim server URL
		url := viper.GetString("server")

		// Read credentials
		uid := viper.GetUint("uid")
		token := ReadCredentials()
		endpoint := fmt.Sprintf("%s/v1/users/%d", url, uid)
		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		userBody := models.JSONUserBody{
			Username:     viper.GetString("username"),
			GivenName:    viper.GetString("firstname"),
			Surname:      viper.GetString("lastname"),
			Email:        viper.GetString("email"),
			SSHPublicKey: viper.GetString("ssh-public-key"),
			MemberOf:     viper.GetString("groups"),
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
	updateUserCmd.Flags().StringP("username", "u", "", "username")
	updateUserCmd.Flags().StringP("firstname", "f", "", "first name")
	updateUserCmd.Flags().StringP("lastname", "l", "", "last name")
	updateUserCmd.Flags().StringP("email", "e", "", "email")
	updateUserCmd.Flags().StringP("ssh-public-key", "k", "", "SSH Public Key")
	updateUserCmd.Flags().StringP("groups", "g", "", "comma-separated list of group names. ")
	updateUserCmd.Flags().Bool("manager", false, "Glim manager account?")
	updateUserCmd.Flags().Bool("readonly", false, "Glim readonly account?")
	updateUserCmd.Flags().Bool("plainuser", false, "Glim plain user account. User can read and modify its own user account information but not its group membership.")
	updateUserCmd.Flags().Bool("replace", false, "replace groups with those specified with -g. Groups are appended to those that the user is a member of by default")
	updateUserCmd.Flags().Bool("remove", false, "remove group membership with those specified with -g.")
	updateUserCmd.Flags().Bool("lock", false, "lock account (cannot log in)")
	updateUserCmd.Flags().Bool("unlock", false, "unlock account (can log in)")
	updateUserCmd.Flags().UintP("uid", "i", 0, "user account id")
	updateUserCmd.MarkFlagRequired("uid")
}
