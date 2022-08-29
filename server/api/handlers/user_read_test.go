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

func TestUserRead(t *testing.T) {
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

	// Log in with admin user and get tokens
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username": "saul", "password": "test"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	h.Login(c, settings)
	response = types.Response{}
	json.Unmarshal(res.Body.Bytes(), &response)
	plainUserToken := response.AccessToken

	everybodyInfo := `[{"uid":1,"username":"admin","name":"","firstname":"LDAP","lastname":"administrator","email":"","ssh_public_key":"","manager":true,"readonly":false,"locked":false},{"uid":2,"username":"search","name":"","firstname":"Read-Only","lastname":"Account","email":"","ssh_public_key":"","manager":false,"readonly":true,"locked":false},{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false},{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}]`

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
			name:             "search user can list everybody's information",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users",
			reqMethod:        http.MethodGet,
			secret:           searchToken,
			expectedBodyJSON: everybodyInfo,
		},
		{
			name:             "manager user can list everybody's information",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users",
			reqMethod:        http.MethodGet,
			secret:           adminToken,
			expectedBodyJSON: everybodyInfo,
		},
		{
			name:             "plain user can only see her own info",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: `{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}`,
		},
		{
			name:             "manager user can see a plainuser account info",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodGet,
			secret:           searchToken,
			expectedBodyJSON: `{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}`,
		},
		{
			name:             "search user can see a plainuser account info",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodGet,
			secret:           searchToken,
			expectedBodyJSON: `{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","manager":false,"readonly":false,"locked":false}`,
		},
		{
			name:             "uid must be an integer",
			expResCode:       http.StatusBadRequest,
			reqURL:           "/v1/users/pepe",
			reqMethod:        http.MethodGet,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"uid param should be a valid integer"}`,
		},
		{
			name:             "search user can't see non-existent account",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/3000",
			reqMethod:        http.MethodGet,
			secret:           searchToken,
			expectedBodyJSON: `{"message":"user not found"}`,
		},
		{
			name:             "manager user can't see non-existent account",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/3000",
			reqMethod:        http.MethodGet,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"user not found"}`,
		},
		{
			name:             "plainuser can't see non-existent account",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/3000",
			reqMethod:        http.MethodGet,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"user not found"}`,
		},
		{
			name:             "search user can't get uid from non-existent username",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/non-existent/uid",
			reqMethod:        http.MethodGet,
			secret:           searchToken,
			expectedBodyJSON: `{"message":"user not found"}`,
		},
		{
			name:             "manager user can't get uid from non-existent account",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/non-existent/uid",
			reqMethod:        http.MethodGet,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"user not found"}`,
		},
		{
			name:             "search user can get uid from username",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/saul/uid",
			reqMethod:        http.MethodGet,
			secret:           searchToken,
			expectedBodyJSON: `{"uid":3}`,
		},
		{
			name:             "manager user can get uid from username",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/saul/uid",
			reqMethod:        http.MethodGet,
			secret:           adminToken,
			expectedBodyJSON: `{"uid":3}`,
		},
		{
			name:             "plainuser user can get uid from username",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/saul/uid",
			reqMethod:        http.MethodGet,
			secret:           plainUserToken,
			expectedBodyJSON: `{"uid":3}`,
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
