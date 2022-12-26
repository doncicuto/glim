package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLogout(t *testing.T) {
	// Setup
	h, e, settings := testSetup(t, false)
	defer testCleanUp()

	// Log in with plain user and get tokens
	accessToken, refreshToken := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:        "Bad request wrong string as token",
			expResCode:  http.StatusBadRequest,
			reqURL:      refreshTokenEndpoint,
			reqBodyJSON: `{"refresh_token": "wrong"}`,
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Logout successful saul",
			expResCode:  http.StatusNoContent,
			reqURL:      refreshTokenEndpoint,
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken),
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Refresh token has been blacklisted",
			expResCode:  http.StatusUnauthorized,
			reqURL:      refreshTokenEndpoint,
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken),
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Refresh token not sent",
			expResCode:  http.StatusBadRequest,
			reqURL:      refreshTokenEndpoint,
			reqBodyJSON: fmt.Sprintf(`{"access_token": "%s"}`, accessToken),
			reqMethod:   http.MethodDelete,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
