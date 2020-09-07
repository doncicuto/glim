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
	"fmt"

	"github.com/muultipla/glim/certs"
	"github.com/spf13/cobra"
)

var (
	organization, hosts, path string
	years                     int
)

// certsCmd represents the certs command
var certsCmd = &cobra.Command{
	Use:   "certs",
	Short: "Create a self-signed CA and client and server certificates to secure communications with Glim",
	Run: func(cmd *cobra.Command, args []string) {
		var config = certs.Config{
			Organization: "Glim Fake Organization, Inc",
			Hosts:        []string{"127.0.0.1", "localhost"},
			OutputPath:   "D:\\Code\\Go\\src\\github.com\\muultipla\\glim\\certs",
			Years:        2,
		}
		err := certs.Generate(&config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(certsCmd)
	certsCmd.Flags().StringVarP(&organization, "organization", "o", "Glim Fake Organization, Inc", "organization name. Default: Glim Fake Organization")
	certsCmd.Flags().StringVarP(&hosts, "addresses", "a", "127.0.0.1, localhost", "comma-separated list of hosts and IP addresses to be added to client/server certificate. Default: 127.0.0.1, localhost")
	certsCmd.Flags().StringVarP(&path, "path", "p", "", "filesystem path for the folder where certificates and private keys files will be stored")
	certsCmd.Flags().IntVarP(&years, "duration", "d", 1, "number of years that we want certificates to be valid. Default: 1")
}
