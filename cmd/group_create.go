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

	"github.com/doncicuto/glim/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/doncicuto/glim/common"
)

func NewGroupCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Glim group",
		PreRun: func(cmd *cobra.Command, _ []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// json output?
			jsonOutput := viper.GetBool("json")

			// Glim server URL
			url := viper.GetString("server")
			endpoint := fmt.Sprintf("%s/v1/groups", url)

			// Get credentials
			token, err := GetCredentials(url)
			if err != nil {
				return err
			}

			// Rest API authentication
			client := RestClient(token.AccessToken)

			resp, err := client.R().
				SetHeader(contentTypeHeader, appJson).
				SetBody(models.JSONGroupBody{
					Name:                      viper.GetString("group"),
					Description:               viper.GetString("description"),
					Members:                   viper.GetString("members"),
					GuacamoleConfigProtocol:   viper.GetString("guacamole-protocol"),
					GuacamoleConfigParameters: viper.GetString("guacamole-parameters"),
				}).
				SetError(&common.APIError{}).
				Post(endpoint)

			if err != nil {
				return fmt.Errorf(common.CantConnectMessage, err)
			}

			if resp.IsError() {
				return fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
			}

			printCmdMessage(cmd, "Group created", jsonOutput)
			return nil
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
	}
	defaultRootPEMFilePath := filepath.Join(homeDir, ".glim", "ca.pem")

	cmd.Flags().StringP("group", "g", "", "our group name")
	cmd.Flags().StringP("description", "d", "", "our group description")
	cmd.Flags().StringP("members", "m", "", "comma-separated list of usernames e.g: admin,tux")

	cmd.MarkFlagRequired("group")
	cmd.PersistentFlags().String("tlscacert", defaultRootPEMFilePath, "trust certs signed only by this CA")
	cmd.PersistentFlags().String("server", "https://127.0.0.1:1323", "glim REST API server address")
	cmd.PersistentFlags().Bool("json", false, "encodes Glim output as json string")
	cmd.Flags().String("guacamole-protocol", "", "Apache Guacamole protocol e.g: vnc")
	cmd.Flags().String("guacamole-parameters", "", "Apache Guacamole config params using a comma-separated list e.g: hostname=localhost,port=5900,password=secret")

	return cmd
}
