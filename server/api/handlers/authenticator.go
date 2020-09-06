package handlers

import (
	"github.com/google/uuid"
	"github.com/muultipla/glim/models"
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
