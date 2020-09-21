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
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var gid string

// groupCmd represents the group command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage Glim groups",
	Run: func(cmd *cobra.Command, args []string) {

		if cmd.Flags().Changed("gid") {
			if gid == "" {
				fmt.Println("Error non-null gid required")
				os.Exit(1)
			}
			id, err := strconv.Atoi(gid)
			if err != nil {
				fmt.Println("Error numeric gid required")
				os.Exit(1)
			}
			getGroup(id)
			os.Exit(0)
		}
		getGroups()
		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(listGroupCmd)
	// groupCmd.AddCommand(newGroupCmd)
	// groupCmd.AddCommand(updateGroupCmd)
	// groupCmd.AddCommand(deleteGroupCmd)
	groupCmd.Flags().StringVarP(&gid, "gid", "i", "", "group id")
}
