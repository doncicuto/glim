package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLogout(t *testing.T) {
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

	// Log in with plain user and get tokens
	accessToken, refreshToken := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:        "Bad request wrong string as token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/v1/login/refresh_token",
			reqBodyJSON: `{"refresh_token": "wrong"}`,
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Logout succesful saul",
			expResCode:  http.StatusNoContent,
			reqURL:      "/v1/login/refresh_token",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken),
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Refresh token has been blacklisted",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/v1/login/refresh_token",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken),
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Refresh token not sent",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/v1/login/refresh_token",
			reqBodyJSON: fmt.Sprintf(`{"access_token": "%s"}`, accessToken),
			reqMethod:   http.MethodDelete,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}