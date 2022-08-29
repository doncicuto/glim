package handlers

import (
	"net/http"
	"testing"
)

func TestUserDelete(t *testing.T) {
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
	adminToken, _ := getUserTokens("admin", h, e, settings)
	searchToken, _ := getUserTokens("search", h, e, settings)

	// Test cases
	testCases := []RestTestCase{
		{
			name:       "wrong token",
			expResCode: http.StatusUnauthorized,
			reqURL:     "/v1/users",
			reqMethod:  http.MethodDelete,
			secret:     "wrong secret",
		},
		{
			name:       "uid not an integer",
			expResCode: http.StatusBadRequest,
			reqURL:     "/v1/users/sdsd",
			reqMethod:  http.MethodDelete,
			secret:     adminToken,
		},
		{
			name:       "uid not found in database",
			expResCode: http.StatusNotFound,
			reqURL:     "/v1/users/100000",
			reqMethod:  http.MethodDelete,
			secret:     adminToken,
		},
		{
			name:       "only manager can delete user",
			expResCode: http.StatusForbidden,
			reqURL:     "/v1/users/5",
			reqMethod:  http.MethodDelete,
			secret:     searchToken,
		},
		{
			name:       "user deleted",
			expResCode: http.StatusNoContent,
			reqURL:     "/v1/users/5",
			reqMethod:  http.MethodDelete,
			secret:     adminToken,
		},
		{
			name:       "already deleted",
			expResCode: http.StatusNotFound,
			reqURL:     "/v1/users/5",
			reqMethod:  http.MethodDelete,
			secret:     adminToken,
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
