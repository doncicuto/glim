package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {

	testCases := []struct {
		name        string
		expResCode  int
		reqURL      string
		reqBodyJSON string
		reqMethod   string
	}{
		{
			name:        "Login succesful saul",
			expResCode:  http.StatusOK,
			reqURL:      "/login",
			reqBodyJSON: `{"username": "saul", "password": "test"}`,
			reqMethod:   http.MethodPost,
		},
		{
			name:        "Login succesful kim",
			expResCode:  http.StatusOK,
			reqURL:      "/login",
			reqBodyJSON: `{"username": "kim", "password": "test"}`,
			reqMethod:   http.MethodPost,
		},
		{
			name:        "Wrong password mike",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/login",
			reqBodyJSON: `{"username": "mike", "password": "boooo"}`,
			reqMethod:   http.MethodPost,
		},
		{
			name:        "User doesn't exist walter",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/login",
			reqBodyJSON: `{"username": "walter", "password": "boooo"}`,
			reqMethod:   http.MethodPost,
		},
		{
			name:        "No JSON body",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/login",
			reqBodyJSON: `""`,
			reqMethod:   http.MethodPost,
		},
	}

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

	for _, tc := range testCases {
		var req *http.Request
		t.Run(tc.name, func(t *testing.T) {
			req = httptest.NewRequest(tc.reqMethod, tc.reqURL, strings.NewReader(tc.reqBodyJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			if err := h.Login(c, settings); err == nil {
				assert.Equal(t, tc.expResCode, res.Code)
			} else {
				he := err.(*echo.HTTPError)
				assert.Equal(t, tc.expResCode, he.Code)
			}
		})
	}
}
