package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIsBlacklisted(t *testing.T) {
	// New SQLite test database
	db, err := newTestDatabase()
	if err != nil {
		t.Fatalf("could not initialize db - %v", err)
	}
	defer removeDatabase()

	// New BadgerDB test key-value storage
	kv, err := newTestKV()
	if err != nil {
		t.Fatalf("could not initialize kv - %v", err)
	}
	defer removeKV()

	settings := testSettings(db, kv)
	e := EchoServer(settings)
	h := &Handler{DB: db, KV: kv}

	// Log in with admin, search and/or plain user and get tokens
	accessToken, refreshToken := getUserTokens("search", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:             "jti not in token",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwicmVhZG9ubHkiOmZhbHNlLCJtYW5hZ2VyIjp0cnVlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.PuDK1A_z108OZb_D4tJfTGUSHRaNHUCQETW7Pf_I2M8",
			expectedBodyJSON: `{"message":"wrong token or missing info in token claims"}`,
		},
		{
			name:       "jti not in KV, not blacklisted",
			expResCode: http.StatusOK,
			reqURL:     "/v1/users/3",
			reqMethod:  http.MethodGet,
			secret:     accessToken,
		},
		{
			name:        "refreshed token successful",
			expResCode:  http.StatusOK,
			reqURL:      "/v1/login/refresh_token",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken),
			reqMethod:   http.MethodPost,
		},
		{
			name:             "jti in KV, access token was blacklisted",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodGet,
			secret:           accessToken,
			expectedBodyJSON: `{"message":"token no longer valid"}`,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}