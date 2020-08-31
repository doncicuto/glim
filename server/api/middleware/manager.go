package middleware

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

//IsManager - TODO comment
func IsManager(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		jwtManager, ok := claims["manager"].(bool)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}

		if !jwtManager {
			return &echo.HTTPError{Code: http.StatusForbidden, Message: "user has no proper permissions"}
		}
		return next(c)
	}
}
