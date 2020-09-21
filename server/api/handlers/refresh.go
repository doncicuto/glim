package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/muultipla/glim/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
	"github.com/muultipla/glim/server/api/auth"
)

// Refresh API tokens
func (h *Handler) Refresh(c echo.Context) error {

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
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not parse token"}
	}

	// Check if JWT token is valid
	if !token.Valid {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "token is not valid"}
	}

	// Extract uid
	uid, ok := claims["uid"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "uid not found in token"}
	}

	// Extract jti
	jti, ok := claims["jti"].(string)
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "jti not found in token"}
	}

	// Extract access token jti
	ajti, ok := claims["ajti"].(string)
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "access jti not found in token"}
	}

	// Extract issued at time claim
	iat, ok := claims["iat"].(float64)
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "iat not found in token"}
	}

	// Check if use of refresh tokens limit has been exceeded
	maxDays := config.MaxDaysWoRelogin()
	refreshLimit := time.Unix(int64(iat), 0).AddDate(0, 0, maxDays).Unix()
	if refreshLimit < time.Now().Unix() {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "refresh token usage without log in exceeded"}
	}

	// Check if user exists
	var dbUser models.User
	if h.DB.Where("id = ?", uid).First(&dbUser).RecordNotFound() {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid uid found in token"}
	}

	// Check if refresh token ID (jti) has been blacklisted
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

	// Blacklist old refresh token
	err = h.KV.Set(jti, "true", time.Second*3600)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store refresh token info"}
	}

	// Blacklist old access token
	err = h.KV.Set(ajti, "true", time.Second*3600)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store refresh token info"}
	}

	// Prepare refresh response

	// Access token expiry time
	expiry := config.AccessTokenExpiry()
	atExpiresIn := time.Second * time.Duration(expiry)
	atExpiresOn := time.Now().Add(atExpiresIn).Unix()

	// Prepare JWT tokens common claims
	cc := jwt.MapClaims{}
	cc["iss"] = "api.glim.server"
	cc["aud"] = "api.glim.server"
	cc["sub"] = "api.glim.client"
	cc["uid"] = dbUser.ID

	// We use request token iat as the iat for new tokens
	// it will be useful to check if we have to login again
	// as the MAX_DAYS_WITHOUT_RELOGIN has been reached
	cc["iat"] = iat

	// Create access claims and token
	tokenID := uuid.New() // token id
	ac := cc              // add common claims to access token claims
	ac["jti"] = tokenID
	ac["exp"] = atExpiresOn
	ac["manager"] = dbUser.Manager
	ac["readonly"] = dbUser.Readonly
	t := jwt.New(jwt.SigningMethodHS256)
	t.Claims = ac
	at, err := t.SignedString([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not create access token"}
	}

	// Add access token to Key-Value store
	err = h.KV.Set(fmt.Sprintf("%s", tokenID), "false", atExpiresIn)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not add access token to key-value store"}
	}

	// Refresh token expiry times
	expiry = config.RefreshTokenExpiry()
	rtExpiresIn := time.Second * time.Duration(expiry)

	// Create response token
	tokenID = uuid.New() // token id
	rc := cc             // add common claims to refresh token claims
	rc["jti"] = tokenID
	rc["exp"] = time.Now().Add(rtExpiresIn).Unix()
	rc["ajti"] = ajti
	t = jwt.New(jwt.SigningMethodHS256)
	t.Claims = rc
	rt, err := t.SignedString([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not create access token"}
	}

	// Add response token to Key-Value store
	err = h.KV.Set(fmt.Sprintf("%s", tokenID), "false", rtExpiresIn)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not add refresh token to key-value store"}
	}

	// Create response with access and refresh tokens
	response := auth.Response{}
	response.AccessToken = at
	response.RefreshToken = rt
	response.TokenType = "Bearer"
	response.ExpiresIn = atExpiresIn.Seconds()
	response.ExpiresOn = atExpiresOn

	return c.JSON(http.StatusOK, response)
}
