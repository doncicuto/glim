package middleware

import (
	"net/http"

	"github.com/dgraph-io/badger"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// IsBlacklisted - TODO comment
func IsBlacklisted(kv *badger.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			user := c.Get("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)

			jti, ok := claims["jti"].(string)
			if !ok {
				return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
			}

			blacklisted := false

			err := kv.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte(jti))
				if err != nil {
					return err
				}
				err = item.Value(func(val []byte) error {
					blacklisted = string(val) == "true"
					return nil
				})
				return err
			})

			if err != nil {
				return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not get stored token info"}
			}

			if blacklisted {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "token no longer valid"}
			}
			return next(c)
		}
	}
}
