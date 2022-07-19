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

package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/doncicuto/glim/config"
	"github.com/doncicuto/glim/server/kv/badgerdb"

	"github.com/doncicuto/glim/server/api"
	"github.com/doncicuto/glim/server/db"
	"github.com/doncicuto/glim/server/ldap"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ()

// serverCmd represents the server command
var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Glim server",
	Run: func(_ *cobra.Command, _ []string) {
		// Get environment variables
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ error getting env, not comming through %v. Exiting now...\n", time.Now().Format(time.RFC3339), err)
			os.Exit(1)
		}

		// Check if API secret is present in env variable
		ok := config.CheckAPISecret()
		if !ok {
			fmt.Printf("%s [Glim] ⇨ could not find required GLIM_API_SECRET environment variable. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		// Check if certificate file path has been specified
		tlscert := viper.GetString("tlscert")
		if tlscert == "" {
			fmt.Printf("%s [Glim] ⇨ Certificate file path cannot be empty. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		if _, err := os.Stat(tlscert); os.IsNotExist(err) {
			fmt.Printf("%s [Glim] ⇨ Certificate file %s cannot be found. Exiting now...\n", time.Now().Format(time.RFC3339), tlscert)
			os.Exit(1)
		}

		// Check if private key file path has been specified
		tlskey := viper.GetString("tlskey")
		if tlskey == "" {
			fmt.Printf("%s [Glim] ⇨ Private key file path cannot be empty. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		if _, err := os.Stat(tlskey); os.IsNotExist(err) {
			fmt.Printf("%s [Glim] ⇨ Private key file %s cannot be found. Exiting now...\n", time.Now().Format(time.RFC3339), tlskey)
			os.Exit(1)
		}

		// Database
		dbName := viper.GetString("db-name")
		database, err := db.Initialize(dbName)
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not connect to database. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		defer func() {
			fmt.Printf("%s [Glim] ⇨ closing connection to database...\n", time.Now().Format(time.RFC3339))
		}()
		fmt.Printf("%s [Glim] ⇨ connected to database...\n", time.Now().Format(time.RFC3339))

		// Key-value store for JWT tokens storage
		// TODO choose between BadgerDB or Redis
		badgerKV := viper.GetString("badgerdb-store")
		blacklist, err := badgerdb.NewBadgerStore(badgerKV)
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not connect to Badger key-value store. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
		defer func() {
			fmt.Printf("%s [Glim] ⇨ closing connection to key-value store...\n", time.Now().Format(time.RFC3339))
			blacklist.Close()
		}()
		fmt.Printf("%s [Glim] ⇨ connected to key-value store...\n", time.Now().Format(time.RFC3339))

		restAddress := viper.GetString("rest-addr")

		// Preparing API server settings
		apiSettings := api.Settings{
			DB:      database,
			KV:      blacklist,
			TLSCert: tlscert,
			TLSKey:  tlskey,
			Address: restAddress,
		}

		ldapAddress := viper.GetString("ldap-addr")

		// Preparing LDAP server settings
		ldapSettings := ldap.Settings{
			DB:      database,
			TLSCert: tlscert,
			TLSKey:  tlskey,
			Address: ldapAddress,
		}

		// get current PID and store it in glim.pid at our tmp directory
		pid := os.Getpid()
		pidFile := filepath.FromSlash(fmt.Sprintf("%s/glim.pid", os.TempDir()))
		err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not store PID in glim.pid. Exiting now...\n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		// Create channels to recieve termination signals and
		// communicate shutdown to servers
		ch := make(chan os.Signal, 1)
		apiShutdownChannel := make(chan bool)
		ldapShutdownChannel := make(chan bool)
		// We listen for the following signals
		signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// Start go routines for both REST and LDAP servers...
		// and add them to a waitgroup
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go api.Server(wg, apiShutdownChannel, apiSettings)
		wg.Add(1)
		go ldap.Server(wg, ldapShutdownChannel, ldapSettings)

		// Blocking wait for server stop signal
		<-ch

		// Send a signal to shutdown both servers
		apiShutdownChannel <- true
		ldapShutdownChannel <- true

		// Wait for both servers to finish
		wg.Wait()
	},
}

func init() {
	homeDir, _ := os.UserHomeDir()
	defaultCertPEMFilePath := filepath.Join(homeDir, ".glim", "server.pem")
	defaultCertKeyFilePath := filepath.Join(homeDir, ".glim", "server.key")

	serverStartCmd.Flags().String("tlscert", defaultCertPEMFilePath, "TLS server certificate path (required)")
	serverStartCmd.Flags().String("tlskey", defaultCertKeyFilePath, "TLS server private key path (required)")
	serverStartCmd.Flags().String("ldap-addr", ":1636", "LDAP server address and port (format: <ip:port>)")
	serverStartCmd.Flags().String("rest-addr", ":1323", "REST API server address and port (format: <ip:port>)")
	serverStartCmd.Flags().String("badgerdb-store", "/tmp/kv", "Directory path for BadgerDB KV store")
	serverStartCmd.Flags().String("db-name", "new.db", "Name of the file containing Glim's database")

	viper.BindPFlags(serverStartCmd.Flags())
}
