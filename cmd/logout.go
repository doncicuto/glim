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

	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout [flags] [SERVER]",
	Short: "Log out from a Glim server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, _ []string) {
		var token *types.Response

		// Read token from file
		token = ReadCredentials()

		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Glim server URL
		url := viper.GetString("server")

		// Logout
		client := RestClient("")

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(fmt.Sprintf(`{"refresh_token":"%s"}`, token.RefreshToken)).
			SetError(&types.APIError{}).
			Delete(fmt.Sprintf("%s/v1/login/refresh_token", url))

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			os.Exit(1)
		}

		// Remove credentials file
		DeleteCredentials()

		fmt.Println("Removing login credentials")
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
	}
	defaultRootPEMFilePath := filepath.Join(homeDir, ".glim", "ca.pem")

	rootCmd.AddCommand(logoutCmd)
	logoutCmd.Flags().String("tlscacert", defaultRootPEMFilePath, "trust certs signed only by this CA")
	logoutCmd.Flags().String("server", "https://127.0.0.1:1323", "glim REST API server address")
}
