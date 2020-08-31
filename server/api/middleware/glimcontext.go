package middleware

import (
	"github.com/dgraph-io/badger"
	"github.com/labstack/echo"
)

//GlimContext - TODO comment
type GlimContext struct {
	Blacklist *badger.DB
	echo.Context
}
