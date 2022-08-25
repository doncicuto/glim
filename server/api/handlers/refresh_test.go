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

func TestRefresh(t *testing.T) {
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
		errMessage  string
	}{
		{
			name:        "Bad request expired token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoyNzcxNTMyODIzLCJpYXQiOjE2NjEyNzM2MjMsImlzcyI6ImFwaS5nbGltLnNlcnZlciIsImp0aSI6ImQ5OGQ0YTA2LTYyOGMtNGNjZC05M2YxLWY5NjNhNmQ0YWU0OSIsIm1hbmFnZXIiOnRydWUsInJlYWRvbmx5IjpmYWxzZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxfQ.1DZfzMDf2jtaVQBFXOmimFpdauuBdoTFcF2N-BNc0sg"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "could not parse token, you may have to log in again",
		},
		{
			name:        "Bad request uid not found in token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoyNzcxNTMyODIzLCJpYXQiOjE2NjEyNzM2MjMsImlzcyI6ImFwaS5nbGltLnNlcnZlciIsImp0aSI6ImQ5OGQ0YTA2LTYyOGMtNGNjZC05M2YxLWY5NjNhNmQ0YWU0OSIsIm1hbmFnZXIiOnRydWUsInJlYWRvbmx5IjpmYWxzZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50In0.ifF_FxTdMbzVoAesbIPayKnm9W9KF3jbiAAIILah7JY"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "uid not found in token",
		},
		{
			name:        "Bad request jti not found in token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoyNzcxNTMyODIzLCJpYXQiOjE2NjEyNzM2MjMsImlzcyI6ImFwaS5nbGltLnNlcnZlciIsIm1hbmFnZXIiOnRydWUsInJlYWRvbmx5IjpmYWxzZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxfQ.mP8ZCNtI6_tx8JnFSzq--ossC9aUUh584vchAfLq0Dw"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "jti not found in token",
		},
		{
			name:        "Bad request access jti not found in token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjI3NzE1MzI4MjMsImlhdCI6MTY2MTI3MzYyMywiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZDk4ZDRhMDYtNjI4Yy00Y2NkLTkzZjEtZjk2M2E2ZDRhZTQ5IiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjF9.5j3Sfwks_4fRtRgcZYVU-sGmBpvClSP9nWxhlKCXCNU"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "access jti not found in token",
		},
		{
			name:        "Bad request access iat not found in token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoyNzcxNTMyODIzLCJpc3MiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJqdGkiOiJkOThkNGEwNi02MjhjLTRjY2QtOTNmMS1mOTYzYTZkNGFlNDkiLCJtYW5hZ2VyIjp0cnVlLCJyZWFkb25seSI6ZmFsc2UsInN1YiI6ImFwaS5nbGltLmNsaWVudCIsInVpZCI6MX0.4TaxwUmi5riq90RRCzvg7CsrBJHxLcbmYalSSmCQULM"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "iat not found in token",
		},
		{
			name:        "Unauthorized expired refresh token",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoyNzcxNTMyODIzLCJpYXQiOjE2NTE1MzI4MjMsImlzcyI6ImFwaS5nbGltLnNlcnZlciIsImp0aSI6ImQ5OGQ0YTA2LTYyOGMtNGNjZC05M2YxLWY5NjNhNmQ0YWU0OSIsIm1hbmFnZXIiOnRydWUsInJlYWRvbmx5IjpmYWxzZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxfQ.T7FIZembax4xD3zozT_9fbEeWsPbJAmG4VkLFl1Fsmk"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "refresh token usage without log in exceeded",
		},
		{
			name:        "Bad request invalid uid found in token",
			expResCode:  http.StatusBadRequest,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: `{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhanRpIjoiOGE1NDM4ZTItMTIxYy00M2U2LWFlZjUtMTU4OWIxMTk2YTBmIiwiYXVkIjoiYXBpLmdsaW0uc2VydmVyIiwiZXhwIjoyNzcxNTMyODIzLCJpYXQiOjE2NjEyNzM2MjMsImlzcyI6ImFwaS5nbGltLnNlcnZlciIsImp0aSI6ImQ5OGQ0YTA2LTYyOGMtNGNjZC05M2YxLWY5NjNhNmQ0YWU0OSIsIm1hbmFnZXIiOnRydWUsInJlYWRvbmx5IjpmYWxzZSwic3ViIjoiYXBpLmdsaW0uY2xpZW50IiwidWlkIjoxMDAwMH0.bo-gc9lUiX0A41_BntUZTLRBtVzoxZqRo6bIrV1Gs4Y"}`,
			reqMethod:   http.MethodPost,
			errMessage:  "invalid uid found in token",
		},
		{
			name:        "Refreshed token successful",
			expResCode:  http.StatusOK,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, response.RefreshToken),
			reqMethod:   http.MethodPost,
			errMessage:  "refresh token usage without log in exceeded",
		},
		{
			name:        "Unauthorized refresh token blacklisted",
			expResCode:  http.StatusUnauthorized,
			reqURL:      "/login/refreshToken",
			reqBodyJSON: fmt.Sprintf(`{"refresh_token": "%s"}`, response.RefreshToken),
			reqMethod:   http.MethodPost,
			errMessage:  "token no longer valid",
		},
	}

	for _, tc := range testCases {
		var req *http.Request
		t.Run(tc.name, func(t *testing.T) {
			req = httptest.NewRequest(tc.reqMethod, tc.reqURL, strings.NewReader(tc.reqBodyJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			if err := h.Refresh(c, settings); err == nil {
				assert.Equal(t, tc.expResCode, res.Code)
			} else {
				he := err.(*echo.HTTPError)
				assert.Equal(t, tc.expResCode, he.Code)
				assert.Equal(t, tc.errMessage, he.Message)
			}
		})
	}
}
