package badgerdb

import (
	"log"
	"runtime"

	"github.com/dgraph-io/badger"
)

// Store implements kv.Store for BadgerDB
type Store struct {
	DB *badger.DB
}

// NewBadgerStore creates a new connection with BadgerDB
func NewBadgerStore() (*badger.DB, error) {
	// Key-value store for JWT tokens storage
	// TODO badgerDB filesystem path should be passed as ENV
	options := badger.DefaultOptions("./server/kv")

	// TODO - Enable or disable badger logging using ENV
	options.Logger = nil

	// In Windows: To avoid "Value log truncate required to run DB. This might result in
	// data loss" we add the options.Truncate = true
	// Reference: https://discuss.dgraph.io/t/lock-issue-on-windows-on-exposed-api/6316.
	if runtime.GOOS == "windows" {
		options.Truncate = true
	}

	db, err := badger.Open(options)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

// Close will terminate a connection with BadgerDB
func (store *Store) Close() error {
	return store.DB.Close()
}
