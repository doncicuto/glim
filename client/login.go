//
// Copyright © 2020 Muultipla Devops <devops@muultipla.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	resty "github.com/go-resty/resty/v2"

	"github.com/spf13/cobra"
)

var username string
var password string
var passwordStdin bool

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login [flags] [SERVER]",
	Short: `Log in to a Glim Server`,
	Long: `Log in to a Glim Server.
If no server is specified, the default is localhost.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		url := "https://127.0.0.1:1323" // TODO - This should not be hardcoded

		if !cmd.Flags().Changed("username") {
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

		if len(args) > 0 {
			url = args[0]
		}

		// Rest API authentication
		client := resty.New()
		// TODO - We should verify server's certificate
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(Credentials{
				Username: username,
				Password: password,
			}).
			SetError(&APIError{}).
			Post(fmt.Sprintf("%s/login", url))

		if err != nil {
			fmt.Printf("Error connecting with Glim: %v\n", err)
			os.Exit(1)
		}

		if resp.IsError() {
			fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
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
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	loginCmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Take the password from stdin")
}
