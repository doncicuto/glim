package handlers

import (
	"net/http"
	"testing"
)

func TestReader(t *testing.T) {
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
	plainUserToken, _ := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:       "invalid token",
			expResCode: http.StatusUnauthorized,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodGet,
			secret:     "wrong secret",
		},
		{
			name:       "uid not found in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodGet,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQifQ.SQ0P6zliTGQiAdTi2DjCDeht0n2FjYdPGV7JgOx0TRY",
		},
		{
			name:       "manager claim not in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodGet,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.j1lc0cK-KtsI5qI6Vpws6mc4RMSwWL-fuobIujGfJYo",
		},
		{
			name:       "readonly claim not in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodGet,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxfQ.eDcXE_IFDAMuvExWiEyQBhJeujL7F7tRrIqKxV6E9rM",
		},
		{
			name:             "plain user can't list everybody's info",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/4",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
		{
			name:             "plainuser user can't get uid from other's username",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/kim/uid",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
		{
			name:             "plainuser user can't get uid from non-existent username",
			expResCode:       http.StatusUnauthorized,
			reqURL:           "/v1/users/walter/uid",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
