package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

// Login - TODO comment
func (h *Handler) Login(c echo.Context) error {
	var authenticator Authenticator
	authenticator.setDB(h.DB)
	authenticator.setKV(h.KV)

	// Parse username and password from body
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	// Authenticate user
	dbUser, err := authenticator.authenticate(*u.Username, *u.Password)
	if err != nil {
		return err
	}

	// Send response
	tokenInfo, err := authenticator.token(dbUser, time.Now().Unix())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tokenInfo)
}

// TokenRefresh - TODO comment
func (h *Handler) TokenRefresh(c echo.Context) error {
	var authenticator Authenticator
	var httpError *echo.HTTPError
	authenticator.setDB(h.DB)
	authenticator.setKV(h.KV)

	// Get refresh token from body
	t := new(AuthTokens)
	if err := c.Bind(t); err != nil {
		return err
	}

	// Get refresh token claims
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(t.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not parse token"}
	}

	if !token.Valid {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "token is not valid"}
	}

	// Extract uid
	uid, ok := claims["uid"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "uid not found in token"}
	}

	// Extract jti
	jti, ok := claims["jti"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "jti not found in token"}
	}

	// Extract access token jti
	ajti, ok := claims["ajti"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "access jti not found in token"}
	}

	// Extract issued at time claim
	iat, ok := claims["iat"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "iat not found in token"}
	}

	// Check if refresh token limit has been exceeded
	maxDays, err := strconv.Atoi(os.Getenv("MAX_DAYS_WITHOUT_RELOGIN"))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not convert maximum days env to int"}
	}
	refreshLimit := time.Unix(int64(iat.(float64)), 0).AddDate(0, 0, maxDays).Unix()
	if refreshLimit < time.Now().Unix() {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "refresh token usage without log in exceeded"}
	}

	// Check if user exists
	dbUser, httpError := authenticator.userFromToken(int32(uid.(float64)))
	if httpError != nil {
		return httpError
	}

	// Check if refresh token has been blacklisted
	if err := authenticator.isBlacklisted(jti.(string)); err != nil {
		return err
	}

	// Blacklist old refresh token
	if err := authenticator.blacklist(jti.(string)); err != nil {
		return err
	}

	// Blacklist old access token, never mind if access token is not stored
	authenticator.blacklist(ajti.(string))

	// Prepare response with new tokens
	tokenInfo, httpError := authenticator.token(dbUser, int64(iat.(float64)))
	if httpError != nil {
		return httpError
	}
	return c.JSON(http.StatusOK, tokenInfo)
}

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
