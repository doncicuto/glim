package middleware

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

//IsReader - TODO comment
func IsReader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		uid := c.Param("uid")
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		jwtID, ok := claims["uid"].(float64)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}
		jwtReadonly, ok := claims["readonly"].(bool)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}
		jwtManager, ok := claims["manager"].(bool)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}
		if jwtManager || !jwtReadonly || (jwtReadonly && uid == fmt.Sprintf("%f", jwtID)) {
			return next(c)
		}
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "user has no proper permissions"}
	}
}
