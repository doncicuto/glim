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
	"strings"

	"github.com/doncicuto/glim/certs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// certsCmd represents the certs command
var certsCmd = &cobra.Command{
	Use:   "certs",
	Short: "Create a self-signed CA and client and server certificates to secure communications with Glim",
	Run: func(cmd *cobra.Command, args []string) {

		var config = certs.Config{}

		// organization cannot be empty
		organization := viper.GetString("organization")
		if organization == "" {
			fmt.Println("Organization name cannot be empty")
			os.Exit(1)
		}
		config.Organization = organization

		// address list cannot be empty
		hosts := strings.Split(viper.GetString("hosts"), ",")
		if len(hosts) == 0 {
			fmt.Println("Please specify a comma-separated list of hosts and/or IP addresses to be added to certificates")
			os.Exit(1)
		}
		config.Hosts = hosts

		path := viper.GetString("path")
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("Could not create selected directory for certificates path")
			os.Exit(1)
		}
		config.OutputPath = path

		// years should be greater than 0
		years := viper.GetInt("years")
		if years < 1 {
			fmt.Println("Certificate should be valid for at least 1 year")
			os.Exit(1)
		}
		config.Years = years

		// create our certificates signed by our fake CA
		err = certs.Generate(&config)
		if err != nil {
			fmt.Printf("Could not generate our certificates. Error: %v\n", err)
		}
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
		os.Exit(1)
	}
	path := fmt.Sprintf("%s/.glim", homeDir)

	rootCmd.AddCommand(certsCmd)
	certsCmd.Flags().String("organization", "Glim Fake Organization, Inc", "organization name")
	certsCmd.Flags().String("hosts", "127.0.0.1, localhost", "comma-separated list of hosts and IP addresses to be added to certificates")
	certsCmd.Flags().String("path", path, "filesystem path where certificates and private keys files will be stored")
	certsCmd.Flags().Int("years", 1, "number of years that we want certificates to be valid.")

	viper.BindPFlags(certsCmd.Flags())
}
