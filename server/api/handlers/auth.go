package handlers

import (
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// Logout - TODO comment
func (h *Handler) Logout(c echo.Context) error {
	var authenticator Authenticator
	authenticator.setDB(h.DB)
	authenticator.setKV(h.KV)

	// Get refresh token from query string
	refreshToken := c.QueryParam("refreshToken")

	// Get access token claims
	rc := make(jwt.MapClaims)
	rt, err := jwt.ParseWithClaims(refreshToken, rc, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("API_SECRET")), nil
	})

	// Extract access token jti
	ajti, ok := rc["ajti"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "use refresh token"}
	}

	if rt.Valid && err == nil {
		if err := authenticator.blacklistIfNeeded(rc); err != nil {
			return err
		}

		// Also blacklist access token
		authenticator.blacklist(ajti.(string))
	}

	return c.NoContent(http.StatusNoContent)
}
