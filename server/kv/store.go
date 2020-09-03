package kv

// Store contains method signatures for common key-value storage functions
// that should provide BadgerDB or Redis
type Store interface {
	// Set a value for a given key
	Set(k string, v interface{}) error
	// Get a value from our key-value store
	Get(k string, v interface{}) (found bool, err error)
	// Delete a value from our key-value store
	Delete(k string) error
	// Close a connection with our key-value store
	Close() error
}
