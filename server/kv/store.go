package kv

import "time"

// Store contains method signatures for common key-value storage functions
// that should provide BadgerDB or Redis
type Store interface {
	// Set a value for a given key
	Set(k string, v string, expiration time.Duration) error
	// Get a value from our key-value store
	Get(k string) (v string, found bool, err error)
	// Close a connection with our key-value store
	Close() error
}
