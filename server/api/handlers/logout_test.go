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
				t.Logf("%s, %v", response.RefreshToken, he.Message)
			}
		})
	}
}

// Old, expired token
// {
//   "token_type": "Bearer",
//   "expires_in": 3600,
//   "expires_on": 1661277223,
//   "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE2NjEyNzcyMjMsImlhdCI6MTY2MTI3MzYyMywiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.TsYg_8JiHvn-sdt8JttY0vCNRWQ-tXboUkDVxUqdFBc",
//   "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoxNjYxNTMyODIzLCJpYXQiOjE2NjEyNzM2MjMsImlzcyI6ImFwaS5nbGltLnNlcnZlciIsImp0aSI6ImQ5OGQ0YTA2LTYyOGMtNGNjZC05M2YxLWY5NjNhNmQ0YWU0OSIsIm1hbmFnZXIiOnRydWUsInJlYWRvbmx5IjpmYWxzZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxfQ.R2XfTbfIdPnqhFLoTvLvE1_RgvmjMG5vXh_fLyjsLaY"
// }
