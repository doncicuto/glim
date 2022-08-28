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

func TestUserCreate(t *testing.T) {
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
			reqURL:     "/v1/users",
			reqMethod:  http.MethodPost,
			secret:     "wrong secret",
		},
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
