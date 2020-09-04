package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
	"github.com/muultipla/glim/server/kv"
)

type jwtSettings struct {
	user            *models.User
	uuid            uuid.UUID
	issuedAt        int64
	expiresIn       float64
	expiresOn       int64
	isAccessToken   bool
	accessTokenUUID uuid.UUID
}

// AuthTokens - TODO comment
type AuthTokens struct {
	TokenType    string  `json:"token_type"`
	Token        string  `json:"access_token"`
	ExpiresIn    float64 `json:"expires_in"`
	ExpiresOn    int64   `json:"expires_on"`
	RefreshToken string  `json:"refresh_token"`
}

// Authenticator - TODO comment
type Authenticator struct {
	db *gorm.DB
	kv kv.Store
}

func (a *Authenticator) setDB(db *gorm.DB) {
	a.db = db
}

func (a *Authenticator) setKV(kv kv.Store) {
	a.kv = kv
}

func (a *Authenticator) authenticate(username, password string) (*models.User, *echo.HTTPError) {
	var dbUser models.User

	// Check if user exists
	if a.db.Where("username = ?", username).First(&dbUser).RecordNotFound() {
		return nil, &echo.HTTPError{Code: http.StatusUnauthorized, Message: "wrong username or password"}
	}

	// Check if passwords match
	if err := models.VerifyPassword(*dbUser.Password, password); err != nil {
		return nil, &echo.HTTPError{Code: http.StatusUnauthorized, Message: "wrong username or password"}
	}

	return &dbUser, nil
}

func (a *Authenticator) addToKV(uuid uuid.UUID) error {
	err := a.kv.Set(fmt.Sprintf("%s", uuid), "false", time.Second*3600)
	return err
}

func jwtToken(settings jwtSettings) (string, error) {
	// Create JWT access token
	t := jwt.New(jwt.SigningMethodHS256)

	// Set JWT claims
	claims := t.Claims.(jwt.MapClaims)
	claims["iss"] = "api.glim.server"
	claims["aud"] = "api.glim.server"
	claims["sub"] = "api.glim.client"
	claims["jti"] = settings.uuid
	claims["exp"] = settings.expiresOn
	claims["iat"] = settings.issuedAt
	claims["uid"] = settings.user.ID
	if settings.isAccessToken {
		claims["manager"] = *settings.user.Manager
		claims["readonly"] = *settings.user.Readonly
	} else {
		claims["ajti"] = settings.accessTokenUUID
	}

	// Generate encoded token and send it as response
	return t.SignedString([]byte(os.Getenv("API_SECRET")))
}

func (a *Authenticator) token(user *models.User, iat int64) (*AuthTokens, *echo.HTTPError) {
	var authToken AuthTokens
	var err error

	// Auth token type is Bearer
	authToken.TokenType = "Bearer"

	// Get expiry time for access token
	atExpiration, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME_SECONDS"))
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not convert expiry time to int"}
	}
	atExpiresIn := time.Second * time.Duration(atExpiration)
	atExpiresOn := time.Now().Add(atExpiresIn).Unix()

	// Create JWT access token
	atUUID := uuid.New()
	authToken.Token, err = jwtToken(jwtSettings{
		user:          user,
		uuid:          atUUID,
		isAccessToken: true,
		issuedAt:      iat,
		expiresIn:     atExpiresIn.Seconds(),
		expiresOn:     atExpiresOn,
	})
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not create JWT token"}
	}
	authToken.ExpiresIn = atExpiresIn.Seconds()
	authToken.ExpiresOn = atExpiresOn

	// Add entry in KV for token UUID
	err = a.addToKV(atUUID)
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not save JWT token"}
	}

	rtExpiration, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_TIME_SECONDS"))
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not convert expiry time to int"}
	}
	rtExpiresIn := time.Second * time.Duration(rtExpiration)
	rtExpiresOn := time.Now().Add(rtExpiresIn).Unix()

	// Create JWT refresh token
	rtUUID := uuid.New()
	authToken.RefreshToken, err = jwtToken(jwtSettings{
		user:            user,
		uuid:            rtUUID,
		accessTokenUUID: atUUID,
		isAccessToken:   false,
		issuedAt:        iat,
		expiresIn:       rtExpiresIn.Seconds(),
		expiresOn:       rtExpiresOn,
	})
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not create JWT token"}
	}

	// Add entry in KV for token UUID
	err = a.addToKV(rtUUID)
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store JWT token"}
	}

	return &authToken, nil
}

func (a *Authenticator) isBlacklisted(jti string) *echo.HTTPError {
	// TODO - Review this assignment
	val, found, err := a.kv.Get(jti)

	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not get stored token info"}
	}

	if found {
		// blacklisted item
		if string(val) == "true" {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "token no longer valid"}
		}
	}

	return nil
}

func (a *Authenticator) userFromToken(uid int32) (*models.User, *echo.HTTPError) {
	var dbUser models.User

	if a.db.Where("id = ?", uid).First(&dbUser).RecordNotFound() {
		return nil, &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid uid found in token"}
	}
	return &dbUser, nil
}

func (a *Authenticator) blacklist(id string) *echo.HTTPError {
	err := a.kv.Set(fmt.Sprintf("%s", id), "true", time.Second*3600)

	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store token info"}
	}
	return nil
}

func (a *Authenticator) blacklistIfNeeded(claims jwt.MapClaims) *echo.HTTPError {
	// Extract jti
	jti, ok := claims["jti"]
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "jti not found in token"}
	}

	// Check if token has been blacklisted
	if err := a.isBlacklisted(jti.(string)); err != nil {
		return err
	}

	// Blacklist access token
	if err := a.blacklist(jti.(string)); err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store token info"}
	}
	return nil
}
