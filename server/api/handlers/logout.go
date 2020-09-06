package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/server/api/auth"
)

// Logout - TODO comment
func (h *Handler) Logout(c echo.Context) error {

	// Get refresh token from body
	tokens := new(auth.Tokens)
	if err := c.Bind(tokens); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not parse token"}
	}

	// Get refresh token claims
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(tokens.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("API_SECRET")), nil
	})

	// Extract access token jti
	ajti, ok := claims["ajti"].(string)
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "use refresh token"}
	}

	if token.Valid && err == nil {

		// Extract jti
		jti, ok := claims["jti"].(string)
		if !ok {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "jti not found in token"}
		}

		// Check if token has been blacklisted
		val, found, err := h.KV.Get(jti)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not get stored token info"}
		}
		if found {
			// blacklisted item?
			if string(val) == "true" {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "token no longer valid"}
			}
		}

		// Blacklist refresh token
		err = h.KV.Set(jti, "true", time.Second*3600)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store refresh token info"}
		}

		// Blacklist access token
		err = h.KV.Set(ajti, "true", time.Second*3600)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store access token info"}
		}
	}

	return c.NoContent(http.StatusNoContent)
}
