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

// Login - TODO comment
func (h *Handler) Login(c echo.Context) error {
	var dbUser models.User

	// Parse username and password from body
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "could not bind json body to user model"}
	}
	username := *u.Username
	password := *u.Password

	// Check if user exists
	if h.DB.Where("username = ?", username).First(&dbUser).RecordNotFound() {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "wrong username or password"}
	}

	// Check if passwords match
	if err := models.VerifyPassword(*dbUser.Password, password); err != nil {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "wrong username or password"}
	}

	// Access token expiry times
	expiry := config.AccessTokenExpiry()
	expiresIn := time.Second * time.Duration(expiry)
	expiresOn := time.Now().Add(expiresIn).Unix()

	// Prepare JWT tokens common claims
	cc := jwt.MapClaims{}
	cc["iss"] = "api.glim.server"
	cc["aud"] = "api.glim.server"
	cc["sub"] = "api.glim.client"
	cc["uid"] = dbUser.ID
	cc["iat"] = time.Now().Unix()
	cc["exp"] = expiresOn

	// Create access claims and token
	ajti := uuid.New() // token id
	ac := cc           // add common claims to access token claims
	ac["jti"] = ajti
	ac["manager"] = dbUser.Manager
	ac["readonly"] = dbUser.Readonly
	t := jwt.New(jwt.SigningMethodHS256)
	t.Claims = ac
	at, err := t.SignedString([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not create access token"}
	}

	// Add access token to Key-Value store
	err = h.KV.Set(fmt.Sprintf("%s", ajti), "false", expiresIn)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not add access token to key-value store"}
	}

	// Refresh token expiry times
	expiry = config.RefreshTokenExpiry()
	expiresIn = time.Second * time.Duration(expiry)
	expiresOn = time.Now().Add(expiresIn).Unix()

	// Create response token
	rjti := uuid.New() // token id
	rc := cc           // add common claims to refresh token claims
	rc["jti"] = rjti
	rc["ajti"] = ajti
	rc["exp"] = expiresOn

	t = jwt.New(jwt.SigningMethodHS256)
	t.Claims = rc
	rt, err := t.SignedString([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not create access token"}
	}

	// Add response token to Key-Value store
	err = h.KV.Set(fmt.Sprintf("%s", rjti), "false", expiresIn)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not add refresh token to key-value store"}
	}

	// Create response with access and refresh tokens
	response := auth.Response{}
	response.AccessToken = at
	response.RefreshToken = rt
	response.TokenType = "Bearer"
	response.ExpiresIn = expiresIn.Seconds()
	response.ExpiresOn = expiresOn

	return c.JSON(http.StatusOK, response)
}
