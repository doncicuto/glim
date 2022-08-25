package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/doncicuto/glim/types"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

	// Log in with user and get tokens
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username": "saul", "password": "test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h.Login(c, settings)

	response := types.Response{}
	json.Unmarshal(res.Body.Bytes(), &response)

	// Test cases
	testCases := []struct {
		name        string
		expResCode  int
		reqURL      string
		reqBodyJSON string
		reqMethod   string
	}{
		{
			name:        "Bad request wrong string as token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "wrong"}`,
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Logout succesful saul",
			expResCode:  http.StatusNoContent,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, response.RefreshToken),
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Refresh token has been blacklisted",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, response.RefreshToken),
			reqMethod:   http.MethodDelete,
		},
		{
			name:        "Refresh token not sent",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: fmt.Sprintf(`{"access_token": "%s"}`, response.AccessToken),
			reqMethod:   http.MethodDelete,
		},
	}

	for _, tc := range testCases {
		var req *http.Request
		t.Run(tc.name, func(t *testing.T) {
			req = httptest.NewRequest(tc.reqMethod, tc.reqURL, strings.NewReader(tc.reqBodyJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			if err := h.Logout(c, settings); err == nil {
				assert.Equal(t, tc.expResCode, res.Code)
			} else {
				he := err.(*echo.HTTPError)
				assert.Equal(t, tc.expResCode, he.Code)
			}
		})
	}
}
