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

	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage Glim user accounts",
	Run: func(cmd *cobra.Command, args []string) {

		_, err := os.Stat(tlscacert)
		if os.IsNotExist(err) {
			fmt.Println("Could not find required CA pem file to validate authority")
			os.Exit(1)
		}

		if cmd.Flags().Changed("uid") {
			getUser(userID, tlscacert)
			os.Exit(0)
		}
		getUsers(tlscacert)
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(listUserCmd)
	userCmd.AddCommand(newUserCmd)
	userCmd.AddCommand(updateUserCmd)
	userCmd.AddCommand(deleteUserCmd)
	userCmd.AddCommand(userPasswdCmd)
	userCmd.Flags().Uint32VarP(&userID, "uid", "i", 0, "user account id")
}
