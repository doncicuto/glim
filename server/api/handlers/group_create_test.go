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

func TestGroupCreate(t *testing.T) {
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
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.j1lc0cK-KtsI5qI6Vpws6mc4RMSwWL-fuobIujGfJYo",
			expectedBodyJSON: `{"message":"wrong token or missing info in token claims"}`,
		},
		{
			name:             "only a valid manager can create a group",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           searchToken,
			expectedBodyJSON: `{"message":"user has no proper permissions"}`,
		},

		{
			name:             "non-existent manager user can't create group",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjEwMDB9.amq5OV7gU7HUrn5YA8sbs2cXMRFeYHTmXm6bhXJ9PDg",
			expectedBodyJSON: `{"message":"wrong user attempting to create group"}`,
		},
		{
			name:             "group name is required",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"required group name"}`,
		},
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
			name:             "group name already exits",
			expResCode:       http.StatusBadRequest,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"name": "devel", "description": "Developers"}`,
			expectedBodyJSON: `{"message":"group already exists"}`,
		},
		{
			name:             "group can be created with existent members",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"name": "lawyers", "description": "Lawyers", "members":"saul,kim"}`,
			expectedBodyJSON: `{"gid":2,"name":"lawyers","description":"Lawyers","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}]}`,
		},
		{
			name:             "group can be created with non-existent members",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"name": "dealers", "description": "Dealers", "members":"walter"}`,
			expectedBodyJSON: `{"gid":3,"name":"dealers","description":"Dealers"}`,
		},
		{
			name:             "group can be created with existent and non-existent members",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/groups",
			reqMethod:        http.MethodPost,
			secret:           adminToken,
			reqBodyJSON:      `{"name": "fixers", "description": "Fixers", "members":"walter,mike"}`,
			expectedBodyJSON: `{"gid":4,"name":"fixers","description":"Fixers","members":[{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}]}`,
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
