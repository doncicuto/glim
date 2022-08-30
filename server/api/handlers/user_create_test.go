package handlers

import (
	"net/http"
	"testing"
)

func TestUserCreate(t *testing.T) {
	// Setup
	h, e, settings := testSetup(t)
	defer testCleanUp()

	// Log in with admin, search and/or plain user and get tokens
	adminToken, _ := getUserTokens("admin", h, e, settings)
	searchToken, _ := getUserTokens("search", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:       "uid in token, not in database",
			expResCode: http.StatusForbidden,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodPost,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjEwMDAwMH0.Tm1ILeeFwO3ZDrMx5tzRN_78iGQtDQSTUFIfYiKpjyA",
		},
		{
			name:       "uid not found in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodPost,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQifQ.SQ0P6zliTGQiAdTi2DjCDeht0n2FjYdPGV7JgOx0TRY",
		},
		{
			name:       "search user cannot create user",
			expResCode: http.StatusForbidden,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodPost,
			secret:     searchToken,
		},
		{
			name:       "empty body not acceptable",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodPost,
			secret:     adminToken,
		},
		{
			name:             "non-existent manager user can't create account info",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users",
			reqMethod:        http.MethodPost,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjEwMDB9.amq5OV7gU7HUrn5YA8sbs2cXMRFeYHTmXm6bhXJ9PDg",
			expectedBodyJSON: `{"message":"wrong user attempting to create user"}`,
		},
		{
			name:        "wrong email",
			expResCode:  http.StatusNotAcceptable,
			reqURL:      "/v1/users",
			reqBodyJSON: `{"username": "jesse", "firstname": "jesse", "lastname": "pickman", "email": "wrongemail"}`,
			reqMethod:   http.MethodPost,
			secret:      adminToken,
		},
		{
			name:             "user created",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users",
			reqBodyJSON:      `{"username": "jesse", "firstname": "jesse", "lastname": "pickman", "email": "chef@example.com"}`,
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			expectedBodyJSON: `{"uid":6,"username":"jesse","name":"jesse pickman","firstname":"jesse","lastname":"pickman","email":"chef@example.com","ssh_public_key":"","manager":false,"readonly":false,"locked":false}`,
		},
		{
			name:             "user created",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users",
			reqBodyJSON:      `{"username": "hank", "firstname": "hank", "lastname": "schraeder", "email": "hank@newmexicopolice.org", "manager": true}`,
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			expectedBodyJSON: `{"uid":7,"username":"hank","name":"hank schraeder","firstname":"hank","lastname":"schraeder","email":"hank@newmexicopolice.org","ssh_public_key":"","manager":true,"readonly":false,"locked":false}`,
		},
		{
			name:        "user already exits",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/v1/users",
			reqBodyJSON: `{"username": "jesse", "firstname": "jesse", "lastname": "pickman", "email": "chef@example.com"}`,
			reqMethod:   http.MethodPost,
			secret:      adminToken,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
