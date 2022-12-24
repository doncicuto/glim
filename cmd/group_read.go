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
	"strings"

	"github.com/doncicuto/glim/models"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/doncicuto/glim/common"
)

func getGIDFromGroupName(client *resty.Client, group string, url string) (uint, error) {
	endpoint := fmt.Sprintf("%s/v1/groups/%s/gid", url, group)
	resp, err := client.R().
		SetHeader(contentTypeHeader, appJson).
		SetResult(models.GroupID{}).
		SetError(&common.APIError{}).
		Get(endpoint)

	if err != nil {
		return 0, fmt.Errorf(common.CantConnectMessage, err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
	}

	result := resp.Result().(*models.GroupID)
	return uint(result.ID), nil
}

func getGroup(cmd *cobra.Command, id uint, jsonOutput bool) error {
	// Glim server URL
	url := viper.GetString("server")
	endpoint := fmt.Sprintf("%s/v1/groups/%d", url, id)

	// Get credentials
	token, err := GetCredentials(url)
	if err != nil {
		return err
	}

	// Rest API authentication
	client := RestClient(token.AccessToken)

	resp, err := client.R().
		SetHeader(contentTypeHeader, appJson).
		SetResult(models.GroupInfo{}).
		SetError(&common.APIError{}).
		Get(endpoint)

	if err != nil {
		return fmt.Errorf(common.CantConnectMessage, err)
	}
	if resp.IsError() {
		return fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
	}

	result := resp.Result().(*models.GroupInfo)

	if jsonOutput {
		encodeGroupToJson(cmd, result)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), descrFormat, "Group:", result.Name)
		fmt.Fprint(cmd.OutOrStdout(), sectionSeparator)
		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-100d\n", " GID:", result.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-100s\n\n", " Description:", result.Description)

		fmt.Fprintf(cmd.OutOrStdout(), "%-15s\n", "Members:")
		fmt.Fprint(cmd.OutOrStdout(), sectionSeparator)
		for _, member := range result.Members {
			fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-100d\n", " UID:", member.ID)
			fmt.Fprintf(cmd.OutOrStdout(), descrFormat, " Username:", member.Username)
			fmt.Fprintf(cmd.OutOrStdout(), "----\n")
		}

		// Support for Apache Guacamole LDAP config schema
		if result.GuacamoleConfigProtocol != "" && result.GuacamoleConfigParameters != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%-15s\n", "Apache Guacamole Configuration:")
			fmt.Fprint(cmd.OutOrStdout(), sectionSeparator)
			fmt.Fprintf(cmd.OutOrStdout(), descrFormat, " Protocol:", result.GuacamoleConfigProtocol)
			parameters := strings.Split(result.GuacamoleConfigParameters, ",")
			if len(parameters) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s\n", " Parameters:")
				for _, parameter := range parameters {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %-17s\n", parameter)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "----\n")
		}
	}
	return nil
}

func getGroups(cmd *cobra.Command, jsonOutput bool) error {
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
		SetResult([]models.GroupInfo{}).
		SetError(&common.APIError{}).
		Get(endpoint)

	if err != nil {
		return fmt.Errorf(common.CantConnectMessage, err)
	}

	if resp.IsError() {
		return fmt.Errorf("%v", resp.Error().(*common.APIError).Message)
	}

	results := resp.Result().(*[]models.GroupInfo)
	if jsonOutput {
		encodeGroupsToJson(cmd, results)
	} else {
		guacamoleEnabled, err := isGuacamoleEnabled(client, url)
		if err != nil {
			return err
		}

		if guacamoleEnabled {
			fmt.Fprintf(cmd.OutOrStdout(), "%-6s %-20s %-35s %-10s %-50s\n",
				"GID",
				"GROUP",
				"DESCRIPTION",
				"GUACAMOLE",
				"MEMBERS",
			)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%-6s %-20s %-35s %-50s\n",
				"GID",
				"GROUP",
				"DESCRIPTION",
				"MEMBERS",
			)
		}

		for _, result := range *results {
			members := "none"
			if len(result.Members) > 0 {
				members = ""
				for i, member := range result.Members {
					if i == len(result.Members)-1 {
						members += member.Username
					} else {
						members += member.Username + ", "
					}
				}
			}

			if guacamoleEnabled {
				fmt.Fprintf(cmd.OutOrStdout(), "%-6d %-20s %-35s %-10t %-50s\n",
					result.ID,
					truncate(result.Name, 20),
					truncate(result.Description, 35),
					result.GuacamoleConfigProtocol != "" && result.GuacamoleConfigParameters != "",
					truncate(members, 50))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%-6d %-20s %-35s %-50s\n",
					result.ID,
					truncate(result.Name, 20),
					truncate(result.Description, 35),
					truncate(members, 50))
			}

		}
	}
	return nil
}

func GetGroupInfo(cmd *cobra.Command) error {
	gid := viper.GetUint("gid")
	group := viper.GetString("group")
	jsonOutput := viper.GetBool("json")
	if gid != 0 {
		if err := getGroup(cmd, gid, jsonOutput); err != nil {
			return err
		}
		return nil
	}
	if group != "" {
		url := viper.GetString("server")
		// Get credentials
		token, err := GetCredentials(url)
		if err != nil {
			return err
		}

		client := RestClient(token.AccessToken)
		gid, err = getGIDFromGroupName(client, group, url)
		if err != nil {
			return err
		}
		if err := getGroup(cmd, gid, jsonOutput); err != nil {
			return err
		}
		return nil
	}
	if err := getGroups(cmd, jsonOutput); err != nil {
		return err
	}
	return nil
}

// ListGroupCmd - TODO comment
func ListGroupCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List Glim groups",
		PreRun: func(cmd *cobra.Command, _ []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return GetGroupInfo(cmd)
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
	}
	defaultRootPEMFilePath := filepath.Join(homeDir, ".glim", "ca.pem")

	cmd.Flags().UintP("gid", "i", 0, "group id")
	cmd.Flags().StringP("group", "g", "", "group name")
	cmd.PersistentFlags().String("tlscacert", defaultRootPEMFilePath, "trust certs signed only by this CA")
	cmd.PersistentFlags().String("server", "https://127.0.0.1:1323", "glim REST API server address")
	cmd.PersistentFlags().Bool("json", false, "encodes Glim output as json string")

	return cmd
}
