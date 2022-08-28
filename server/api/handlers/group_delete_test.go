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
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestGroupDelete(t *testing.T) {
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

	// Log in with admin user and get tokens
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username": "admin", "password": "test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h.Login(c, settings)
	response := types.Response{}
	json.Unmarshal(res.Body.Bytes(), &response)
	adminToken := response.AccessToken

	// Log in with admin user and get tokens
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username": "search", "password": "test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	h.Login(c, settings)
	response = types.Response{}
	json.Unmarshal(res.Body.Bytes(), &response)
	searchToken := response.AccessToken

	// Test cases
	testCases := []struct {
		name             string
		expResCode       int
		reqURL           string
		reqBodyJSON      string
		reqMethod        string
		secret           string
		expectedBodyJSON string
	}{
		{
			name:       "invalid token",
			expResCode: http.StatusUnauthorized,
			reqURL:     "/v1/groups",
			reqMethod:  http.MethodPost,
			secret:     "wrong secret",
		},
		{
			name:             "manager claim not in token",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/groups/1",
			reqMethod:        http.MethodDelete,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.j1lc0cK-KtsI5qI6Vpws6mc4RMSwWL-fuobIujGfJYo",
			expectedBodyJSON: `{"message":"wrong token or missing info in token claims"}`,
		},
		{
			name:             "only a valid manager can delete a group",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/groups/1",
			reqMethod:        http.MethodDelete,
			secret:           searchToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},
		{
			name:             "gid param should be a valid integer",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/groups/wrong",
			reqMethod:        http.MethodDelete,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"gid param should be a valid integer"}`,
		},
		{
			name:             "group not found",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/groups/1000",
			reqMethod:        http.MethodDelete,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"group not found"}`,
		},
		{
			name:             "create group",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"name": "devel", "description": "Developers"}`,
			expectedBodyJSON: `{"gid":1,"name":"devel","description":"Developers"}`,
		},
		{
			name:       "group can be deleted",
			expResCode: http.StatusNoContent,
			reqURL:     "/v1/groups/1",
			reqMethod:  http.MethodDelete,
			secret:     adminToken,
		},
	}

	for _, tc := range testCases {
		var req *http.Request
		t.Run(tc.name, func(t *testing.T) {
			req = httptest.NewRequest(tc.reqMethod, tc.reqURL, strings.NewReader(tc.reqBodyJSON))
			req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %v", tc.secret))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			assert.Equal(t, tc.expResCode, res.Code)
			if tc.expectedBodyJSON != "" {
				assert.Equal(t, tc.expectedBodyJSON, strings.TrimSuffix(res.Body.String(), "\n"))
			}
		})
	}
}
