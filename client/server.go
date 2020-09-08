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
	"sync"
	"time"

	"github.com/muultipla/glim/config"
	"github.com/muultipla/glim/server/kv/badgerdb"

	"github.com/joho/godotenv"
	"github.com/muultipla/glim/server/api"
	"github.com/muultipla/glim/server/db"
	"github.com/muultipla/glim/server/ldap"
	"github.com/spf13/cobra"
)

var (
	tlscert, tlskey string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage a Glim server",
	Run: func(cmd *cobra.Command, args []string) {
		// Get environment variables
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ error getting env, not comming through %v. Exiting now...\n", time.Now().Format(time.RFC3339), err)
			os.Exit(1)
		}

		// Check if API secret is present in env variable
		ok := config.CheckAPISecret()
		if !ok {
			fmt.Printf("%s [Glim] ⇨ could not find required API_SECRET environment variable. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		// Check if certificate file path has been specified
		if tlscert == "" {
			fmt.Printf("%s [Glim] ⇨ Certificate file path cannot be empty. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		if _, err := os.Stat(tlscert); os.IsNotExist(err) {
			fmt.Printf("%s [Glim] ⇨ Certificate file cannot be found. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		// Check if private key file path has been specified
		if tlskey == "" {
			fmt.Printf("%s [Glim] ⇨ Private key file path cannot be empty. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		if _, err := os.Stat(tlskey); os.IsNotExist(err) {
			fmt.Printf("%s [Glim] ⇨ Private key file cannot be found. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		// Database
		database, err := db.Initialize()
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not connect to database. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		defer database.Close()
		fmt.Printf("%s [Glim] ⇨ connected to database...\n", time.Now().Format(time.RFC3339))

		// Key-value store for JWT tokens storage
		// TODO choose between BadgerDB or Redis
		blacklist, err := badgerdb.NewBadgerStore()
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not connect to Badger key-value store. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		defer blacklist.Close()
		fmt.Printf("%s [Glim] ⇨ connected to key-value store...\n", time.Now().Format(time.RFC3339))

		// Preparing API server settings
		apiSettings := api.Settings{
			DB:      database,
			KV:      blacklist,
			TLSCert: tlscert,
			TLSKey:  tlskey,
		}

		// Preparing LDAP server settings
		ldapSettings := ldap.Settings{
			DB:      database,
			TLSCert: tlscert,
			TLSKey:  tlskey,
		}

		// Start go routines for both REST and LDAP servers...
		var wg sync.WaitGroup
		wg.Add(1)
		go api.Server(&wg, apiSettings)
		wg.Add(2)
		go ldap.Server(&wg, ldapSettings)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&tlscert, "tlscert", "", "TLS server certificate path (required)")
	serverCmd.Flags().StringVar(&tlskey, "tlskey", "", "TLS server private key path (required)")

	// Mark required flags
	cobra.MarkFlagRequired(serverCmd.Flags(), "tlscert")
	cobra.MarkFlagRequired(serverCmd.Flags(), "tlskey")
}
