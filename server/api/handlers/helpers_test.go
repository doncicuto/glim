package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/doncicuto/glim/common"
	"github.com/doncicuto/glim/server/db"
	"github.com/doncicuto/glim/server/kv/badgerdb"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type RestTestCase struct {
	name             string
	expResCode       int
	reqURL           string
	reqBodyJSON      string
	reqMethod        string
	secret           string
	expectedBodyJSON string
}

func newTestDatabase() (*gorm.DB, error) {
	var dbInit = common.DBInit{
		AdminPasswd:   "test",
		SearchPasswd:  "test",
		Users:         "saul,kim,mike",
		DefaultPasswd: "test",
		UseSqlite:     true,
	}
	sqlLog := false
	return db.Initialize("/tmp/test.db", sqlLog, dbInit)
}

func newTestKV() (badgerdb.Store, error) {
	return badgerdb.NewBadgerStore("/tmp/kv")
}

func testSettings(db *gorm.DB, kv common.Store) common.APISettings {
	return common.APISettings{
		DB:                 db,
		KV:                 kv,
		TLSCert:            "",
		TLSKey:             "",
		Address:            "",
		APISecret:          "secret",
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 259200,
		MaxDaysWoRelogin:   7,
	}
}

func testSetup(t *testing.T, guacamole bool) (*Handler, *echo.Echo, common.APISettings) {
	// New SQLite test database
	db, err := newTestDatabase()
	if err != nil {
		t.Fatalf("could not initialize db - %v", err)
	}

	// New BadgerDB test key-value storage
	kv, err := newTestKV()
	if err != nil {
		t.Fatalf("could not initialize kv - %v", err)
	}

	settings := testSettings(db, kv)
	settings.Guacamole = guacamole
	e := EchoServer(settings)
	h := &Handler{DB: db, KV: kv}

	return h, e, settings
}

func testCleanUp() {
	removeDatabase()
	removeKV()
}

func runTests(t *testing.T, tc RestTestCase, e *echo.Echo) {
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

func getUserTokens(username string, h *Handler, e *echo.Echo, settings common.APISettings) (string, string) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(fmt.Sprintf(`{"username": "%s", "password": "test"}`, username)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h.Login(c, settings)
	tokenAuth := common.TokenAuthentication{}
	json.Unmarshal(res.Body.Bytes(), &tokenAuth)
	return tokenAuth.AccessToken, tokenAuth.RefreshToken
}

func removeDatabase() {
	os.Remove("/tmp/test.db")
}

func removeKV() {
	os.RemoveAll("/tmp/kv")
}
