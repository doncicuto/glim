package handlers

import (
	"github.com/jinzhu/gorm"
	"github.com/muultipla/glim/server/kv"
)

//Handler - TODO comment
type Handler struct {
	DB *gorm.DB
	KV kv.Store
}
