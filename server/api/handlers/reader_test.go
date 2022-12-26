package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/doncicuto/glim/common"
)

func TestReader(t *testing.T) {
	// Setup
	h, e, settings := testSetup(t, false)
	defer testCleanUp()

	// Log in with admin, search and/or plain user and get tokens
	plainUserToken, _ := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:             "plain user can't list everybody's info",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/4",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
		{
			name:             "plainuser user can't get uid from other's username",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/kim/uid",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
		{
			name:             "plainuser user can't get uid from non-existent username",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/walter/uid",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
