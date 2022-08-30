package handlers

import (
	"net/http"
	"testing"
)

func TestGroupAddMembers(t *testing.T) {
	// Setup
	h, e, settings := testSetup(t)
	defer testCleanUp()

	// Log in with admin, search and/or plain user and get tokens
	adminToken, _ := getUserTokens("admin", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:             "group can be created without members",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"name": "devel", "description": "Developers"}`,
			expectedBodyJSON: `{"gid":1,"name":"devel","description":"Developers"}`,
		},
		{
			name:             "group id is required",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/groups//members",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"required group gid"}`,
		},
		{
			name:             "group not found",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/groups/100/members",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"group not found"}`,
		},
		{
			name:             "group id is required",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/groups/1/members",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"members": "saul,kim,mike"}`,
			expectedBodyJSON: `{"gid":1,"name":"devel","description":"Developers","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false},{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}]}`,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
