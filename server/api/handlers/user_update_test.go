package handlers

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4/middleware"
)

func TestUserUpdate(t *testing.T) {
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
	e.Pre(middleware.JWT([]byte(settings.APISecret)))
	h := &Handler{DB: db, KV: kv}

	// Log in with admin, search and/or plain user and get tokens
	adminToken, _ := getUserTokens("admin", h, e, settings)
	searchToken, _ := getUserTokens("search", h, e, settings)
	plainUserToken, _ := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:       "invalid token",
			expResCode: http.StatusUnauthorized,
			reqURL:     "/v1/users/3",
			reqMethod:  http.MethodPut,
			secret:     "wrong secret",
		},
		{
			name:       "uid not found in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users/3",
			reqMethod:  http.MethodPut,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQifQ.SQ0P6zliTGQiAdTi2DjCDeht0n2FjYdPGV7JgOx0TRY",
		},
		{
			name:       "manager claim not in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users/3",
			reqMethod:  http.MethodPut,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.j1lc0cK-KtsI5qI6Vpws6mc4RMSwWL-fuobIujGfJYo",
		},
		{
			name:       "readonly claim not in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users/3",
			reqMethod:  http.MethodPut,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxfQ.eDcXE_IFDAMuvExWiEyQBhJeujL7F7tRrIqKxV6E9rM",
		},
		{
			name:             "search user can't update accounts",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           searchToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
		{
			name:             "plainuser can't update other's accounts",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/4",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
		{
			name:             "non-existent manager user can't update account info",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/5",
			reqMethod:        http.MethodPut,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjEwMDB9.amq5OV7gU7HUrn5YA8sbs2cXMRFeYHTmXm6bhXJ9PDg",
			expectedBodyJSON: `{"message":"wrong user attempting to update account"}`,
		},
		{
			name:             "uid must be an integer",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/wrong",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"uid param should be a valid integer"}`,
		},
		{
			name:             "non-existent accounts can't be updated",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/3000",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"user not found"}`,
		},
		{
			name:             "only managers can update a username",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"username":"walter"}`,
			expectedBodyJSON: `{"message":"only managers can update the username"}`,
		},
		{
			name:             "only managers can update a username",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			reqBodyJSON:      `{"username":"kim"}`,
			expectedBodyJSON: `{"message":"username cannot be duplicated"}`,
		},
		{
			name:             "email must be valid",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"email":"wrong"}`,
			expectedBodyJSON: `{"message":"invalid email"}`,
		},
		{
			name:             "only managers can update manager status",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"manager":true}`,
			expectedBodyJSON: `{"message":"only managers can update manager status"}`,
		},
		{
			name:             "only managers can update readonly status",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"readonly":true}`,
			expectedBodyJSON: `{"message":"only managers can update readonly status"}`,
		},
		{
			name:             "only managers can update locked status",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"locked":true}`,
			expectedBodyJSON: `{"message":"only managers can update locked status"}`,
		},
		{
			name:             "plainuser can update her acount",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"firstname":"saul","lastname":"goodman","email":"new@email.com","ssh_public_key":"key"}`,
			expectedBodyJSON: `{"uid":3,"username":"saul","name":"saul goodman","firstname":"saul","lastname":"goodman","email":"new@email.com","ssh_public_key":"key","manager":false,"readonly":false,"locked":false}`,
		},
		{
			name:             "plainuser can update her acount",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			reqBodyJSON:      `{"firstname":"saul","lastname":"goodman","email":"new@email.com","ssh_public_key":"key"}`,
			expectedBodyJSON: `{"uid":3,"username":"saul","name":"saul goodman","firstname":"saul","lastname":"goodman","email":"new@email.com","ssh_public_key":"key","manager":false,"readonly":false,"locked":false}`,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
