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
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/joho/godotenv"
	"github.com/muultipla/glim/server/api"
	"github.com/muultipla/glim/server/db"
	"github.com/muultipla/glim/server/ldap"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage a Glim server",
	Run: func(cmd *cobra.Command, args []string) {
		// Get environment variables
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error getting env, not comming through %v", err)
		}

		// Database
		database, err := db.Initialize()
		if err != nil {
			log.Fatal("could not connect to database, exiting now...")
		}
		defer database.Close()
		fmt.Printf("%s [Glim] ⇨ connected to database...\n", time.Now().Format(time.RFC3339))

		// Key-value store for JWT tokens storage
		options := badger.DefaultOptions("./server/kv")

		// TODO - Enable or disable badger logging
		options.Logger = nil

		// In Windows: To avoid "Value log truncate required to run DB. This might result in data loss" we add the options.Truncate = true
		// Reference: https://discuss.dgraph.io/t/lock-issue-on-windows-on-exposed-api/6316.
		if runtime.GOOS == "windows" {
			options.Truncate = true
		}

		blacklist, err := badger.Open(options)
		if err != nil {
			log.Fatal(err)
		}
		defer blacklist.Close()
		fmt.Printf("%s [Glim] ⇨ connected to key-value store...\n", time.Now().Format(time.RFC3339))

		// Start go routines for both REST and LDAP servers...
		var wg sync.WaitGroup
		wg.Add(1)
		go api.Server(&wg, database, blacklist)
		wg.Add(2)
		go ldap.Server(&wg, database)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
