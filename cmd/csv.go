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

	"github.com/doncicuto/glim/models"
	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func readUsersFromCSV() []*models.User {
	// Read and open file
	file := viper.GetString("file")
	csvFile, err := os.Open(file)
	if err != nil {
		fmt.Println("can't open CSV file")
		os.Exit(1)
	}
	defer csvFile.Close()

	// Try to unmarshal CSV file usin gocsv
	users := []*models.User{}
	if err := gocsv.UnmarshalFile(csvFile, &users); err != nil { // Load clients from file
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return users
}

// importCmd represents the import command
var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Manage users and groups with CSV files",
	PreRun: func(cmd *cobra.Command, _ []string) {
		viper.BindPFlags(cmd.Flags())
	},
}

func init() {
	rootCmd.AddCommand(csvCmd)
	csvCmd.AddCommand(csvUsersCmd)
}
