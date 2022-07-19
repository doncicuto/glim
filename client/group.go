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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// groupCmd represents the group command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage Glim groups",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, args []string) {
		gid := viper.GetUint("gid")
		if gid != 0 {
			getGroup(gid)
			os.Exit(0)
		}
		getGroups()
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(listGroupCmd)
	groupCmd.AddCommand(newGroupCmd)
	groupCmd.AddCommand(updateGroupCmd)
	groupCmd.AddCommand(deleteGroupCmd)
	groupCmd.Flags().UintP("gid", "g", 0, "group id")
}
