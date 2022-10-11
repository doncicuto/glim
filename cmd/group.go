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
	RunE: func(cmd *cobra.Command, _ []string) error {
		return GetGroupInfo(cmd)
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
	groupCmd.PersistentFlags().Bool("json", false, "encodes Glim output as json string")
	groupCmd.AddCommand(ListGroupCmd())
	groupCmd.AddCommand(NewGroupCmd())
	groupCmd.AddCommand(UpdateGroupCmd())
	groupCmd.AddCommand(DeleteGroupCmd())
	groupCmd.Flags().UintP("gid", "i", 0, "group id")
	groupCmd.Flags().StringP("group", "g", "", "group name")
}
