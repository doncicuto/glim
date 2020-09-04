package middleware

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/server/kv"
)

// IsBlacklisted - TODO comment
func IsBlacklisted(kv kv.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			user := c.Get("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)

			jti, ok := claims["jti"].(string)
			if !ok {
				return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
			}

			// TODO - Review this assignment
			val, found, err := kv.Get(jti)

			if err != nil {
				return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not get stored token info"}
			}

			if found {
				// blacklisted item
				if string(val) == "true" {
					return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "token no longer valid"}
				}
			}

			return next(c)
		}
	}
}
