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
	"crypto/tls"
	"fmt"
	"os"

	resty "github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/server/api/auth"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout [flags] [URL]",
	Short: "Log out from a Glim server",
	Long:  "Log out from a Glim server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var token *auth.Response

		// Read token from file
		token = ReadCredentials()

		// Check expiration
		if NeedsRefresh(token) {
			Refresh(token.RefreshToken)
			token = ReadCredentials()
		}

		// Logout
		client := resty.New()
		// TODO - We should verify server's certificate
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(fmt.Sprintf(`{"refresh_token":"%s"}`, token.RefreshToken)).
			SetError(&APIError{}).
			Delete(fmt.Sprintf("%s/login/refresh_token", url))

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
			os.Exit(1)
		}

		// Remove credentials file
		DeleteCredentials()

		fmt.Println("Removing login credentials")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
