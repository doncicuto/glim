/*
Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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
