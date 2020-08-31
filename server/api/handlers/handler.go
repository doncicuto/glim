package handlers

import (
	"github.com/dgraph-io/badger"
	"github.com/jinzhu/gorm"
)

//Handler - TODO comment
type Handler struct {
	DB *gorm.DB
	KV *badger.DB
}
