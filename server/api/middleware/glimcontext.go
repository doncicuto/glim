package middleware

import (
	"github.com/labstack/echo"
	"github.com/muultipla/glim/server/kv"
)

//GlimContext - TODO comment
type GlimContext struct {
	Blacklist *kv.Store
	echo.Context
}
