package handlers

import (
	"net/http"
	"testing"
)

func TestIsManager(t *testing.T) {
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
	searchToken, _ := getUserTokens("search", h, e, settings)
	plainUserToken, _ := getUserTokens("saul", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:       "manager claim not in token",
			expResCode: http.StatusNotAcceptable,
			reqURL:     "/v1/users/3",
			reqMethod:  http.MethodPut,
			secret:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.j1lc0cK-KtsI5qI6Vpws6mc4RMSwWL-fuobIujGfJYo",
		},
		{
			name:             "readonly user is not manager",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           searchToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
		{
			name:             "plain user is not manager",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/4",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
