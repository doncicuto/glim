package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/doncicuto/glim/common"
)

func TestIsManager(t *testing.T) {
	// Setup
	h, e, settings := testSetup(t, false)
	defer testCleanUp()

	// Log in with admin, search and/or plain user and get tokens
	searchToken, _ := getUserTokens("search", h, e, settings)
	plainUserToken, _ := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:       "manager claim not in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     usersEndpoint,
			reqMethod:  http.MethodPost,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.j1lc0cK-KtsI5qI6Vpws6mc4RMSwWL-fuobIujGfJYo",
		},
		{
			name:             "readonly user is not manager",
			expResCode:       http.StatusForbidden,
			reqURL:           usersEndpoint,
			reqMethod:        http.MethodPost,
			secret:           searchToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
		{
			name:             "plain user is not manager",
			expResCode:       http.StatusForbidden,
			reqURL:           usersEndpoint,
			reqMethod:        http.MethodPost,
			secret:           plainUserToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
