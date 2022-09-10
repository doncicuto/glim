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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// groupCmd represents the group command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage Glim groups",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(_ *cobra.Command, _ []string) {
		GetGroupInfo()
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
	}
	defaultRootPEMFilePath := filepath.Join(homeDir, ".glim", "ca.pem")

	rootCmd.AddCommand(groupCmd)
	groupCmd.PersistentFlags().String("tlscacert", defaultRootPEMFilePath, "trust certs signed only by this CA")
	groupCmd.PersistentFlags().String("server", "https://127.0.0.1:1323", "glim REST API server address")
	groupCmd.AddCommand(listGroupCmd)
	groupCmd.AddCommand(newGroupCmd)
	groupCmd.AddCommand(updateGroupCmd)
	groupCmd.AddCommand(deleteGroupCmd)
	groupCmd.Flags().UintP("gid", "i", 0, "group id")
	groupCmd.Flags().StringP("group", "g", "", "group name")
}
