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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/doncicuto/glim/common"
)

func LogoutCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "logout [flags] [SERVER]",
		Short: "Log out from a Glim server",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, _ []string) error {
			var token *common.TokenAuthentication
			url := viper.GetString("server")

			// Get credentials
			token, err := GetCredentials(url)
			if err != nil {
				return err
			}

			// Logout
			client := RestClient("")

			resp, clientError := client.R().
				SetHeader(contentTypeHeader, appJson).
				SetBody(fmt.Sprintf(`{"refresh_token":"%s"}`, token.RefreshToken)).
				SetError(&common.APIError{}).
				Delete(fmt.Sprintf("%s/v1/login/refresh_token", url))

			if err := checkAPICallResponse(clientError, resp); err != nil {
				return err
			}

			// Remove credentials file
			err = DeleteCredentials()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removing login credentials\n")
			return nil
		},
	}

	cmd.Flags().String("tlscacert", getDefaultRootPEMFilePath(), "trust certs signed only by this CA")
	cmd.Flags().String("server", "https://127.0.0.1:1323", "glim REST API server address")

	return cmd
}

func init() {
	logoutCmd := LogoutCmd()
	rootCmd.AddCommand(logoutCmd)
}
