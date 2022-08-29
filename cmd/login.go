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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/doncicuto/glim/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: `Log in to a Glim Server`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, _ []string) {

		username := viper.GetString("username")
		password := viper.GetString("password")
		passwordStdin := viper.GetBool("password-stdin")

		if username == "" {
			username = prompter.Prompt("Username", "")
			if username == "" {
				fmt.Println("Error non-null username required")
				os.Exit(1)
			}
		}

		if !cmd.Flags().Changed("password") {
			if !passwordStdin {
				password = prompter.Password("Password")
				if password == "" {
					fmt.Println("Error password required")
					os.Exit(1)
				}
			}
		} else {
			fmt.Println("WARNING! Using --password via the CLI is insecure. Use --password-stdin.")
		}

		if passwordStdin {
			if password != "" {
				fmt.Println("--password and --password-stdin are mutually exclusive")
				os.Exit(1)
			} else {
				// Reference: https://flaviocopes.com/go-shell-pipes/
				info, err := os.Stdin.Stat()
				if err != nil {
					fmt.Println("Error reading from stdin")
					os.Exit(1)
				}

				if info.Mode()&os.ModeCharDevice != 0 {
					fmt.Println("Error expecting password from stdin using a pipe")
					os.Exit(1)
				}

				reader := bufio.NewReader(os.Stdin)
				var output []rune

				for {
					input, _, err := reader.ReadRune()
					if err != nil && err == io.EOF {
						break
					}
					output = append(output, input)
				}

				password = strings.TrimSuffix(string(output), "\n")
			}
		}

		// Glim server URL
		url := viper.GetString("server")

		// Rest API authentication
		client := RestClient("")

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(types.Credentials{
				Username: username,
				Password: password,
			}).
			SetError(&types.APIError{}).
			Post(fmt.Sprintf("%s/v1/login", url))

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*types.APIError).Message)
			os.Exit(1)
		}

		// Authenticated, let's store tokens in $HOME/.glim/accessToken.json
		tokenFile, err := AuthTokenPath()
		if err != nil {
			fmt.Printf("%v", err)
		}

		f, err := os.OpenFile(*tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			fmt.Printf("Could not create file to store auth token: %v\n", err)
		}
		defer f.Close()

		if _, err := f.WriteString(resp.String()); err != nil {
			fmt.Printf("Could not store credentials in our local fs: %v\n", err)
		}

		fmt.Println("Login Succeeded")
	},
}

func init() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
	}
	defaultRootPEMFilePath := filepath.Join(homeDir, ".glim", "ca.pem")

	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("tlscacert", defaultRootPEMFilePath, "trust certs signed only by this CA")
	loginCmd.Flags().String("server", "https://127.0.0.1:1323", "glim REST API server address")
	loginCmd.Flags().StringP("username", "u", "", "Username")
	loginCmd.Flags().StringP("password", "p", "", "Password")
	loginCmd.Flags().Bool("password-stdin", false, "Take the password from stdin")

	viper.BindPFlags(loginCmd.Flags())
}
