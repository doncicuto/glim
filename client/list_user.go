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

	"github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/models"
	"github.com/spf13/cobra"
)

var url = "http://127.0.0.1:1323" // TODO - this should not be hardcoded

func getUser(id int) {
	endpoint := fmt.Sprintf("%s/users/%d", url, id)
	// Read credentials
	token := ReadCredentials()

	// Check expiration
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	// Rest API authentication
	client := resty.New()
	client.SetAuthToken(token.AccessToken)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(models.User{}).
		SetError(&APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
		os.Exit(1)
	}

	fmt.Printf("%-10s %-20s %-35s %-35s %-8s %-8s\n",
		"UID",
		"USERNAME",
		"FULLNAME",
		"EMAIL",
		"MANAGER",
		"READONLY",
	)

	result := resp.Result().(*models.User)
	fmt.Printf("%-10d %-20s %-35s %-35s %-8v %-8v\n",
		result.ID,
		*result.Username,
		*result.Fullname,
		*result.Email,
		*result.Manager,
		*result.Readonly,
	)
}

func getUsers() {
	// Read credentials
	token := ReadCredentials()
	endpoint := fmt.Sprintf("%s/users", url)
	// Check expiration
	if NeedsRefresh(token) {
		Refresh(token.RefreshToken)
		token = ReadCredentials()
	}

	// Rest API authentication
	client := resty.New()
	client.SetAuthToken(token.AccessToken)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult([]models.User{}).
		SetError(&APIError{}).
		Get(endpoint)

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
		os.Exit(1)
	}

	fmt.Printf("%-10s %-20s %-35s %-35s %-8s %-8s\n",
		"UID",
		"USERNAME",
		"FULLNAME",
		"EMAIL",
		"MANAGER",
		"READONLY",
	)

	results := resp.Result().(*[]models.User)
	for _, result := range *results {
		fmt.Printf("%-10d %-20s %-35s %-35s %-8v %-8v\n",
			result.ID,
			*result.Username,
			*result.Fullname,
			*result.Email,
			*result.Manager,
			*result.Readonly,
		)
	}
}

// ListUserCmd - TODO comment
var listUserCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Glim user accounts",
	Run: func(cmd *cobra.Command, args []string) {
		getUsers()
	},
}
