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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage Glim user accounts",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {

		tlscacert := viper.GetString("tlscacert")
		_, err := os.Stat(tlscacert)
		if os.IsNotExist(err) {
			fmt.Println("Could not find required CA pem file to validate authority")
			os.Exit(1)
		}

		GetUserInfo()
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(listUserCmd)
	userCmd.AddCommand(newUserCmd)
	userCmd.AddCommand(updateUserCmd)
	userCmd.AddCommand(deleteUserCmd)
	userCmd.AddCommand(userPasswdCmd)
	userCmd.Flags().UintP("uid", "i", 0, "user account id")
	userCmd.Flags().StringP("username", "u", "", "username")
}
