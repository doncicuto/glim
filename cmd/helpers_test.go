package cmd

import (
	"os"
	"testing"

	"github.com/doncicuto/glim/server/api/handlers"
	"github.com/doncicuto/glim/server/db"
	"github.com/doncicuto/glim/server/kv/badgerdb"
	"github.com/doncicuto/glim/types"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

func newTestDatabase() (*gorm.DB, error) {
	var dbInit = types.DBInit{
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

func testSettings(db *gorm.DB, kv types.Store) types.APISettings {
	return types.APISettings{
		DB:                 db,
		KV:                 kv,
		TLSCert:            "",
		TLSKey:             "",
		Address:            "127.0.0.1:1323",
		APISecret:          "secret",
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 259200,
		MaxDaysWoRelogin:   7,
	}
}

func testSetup(t *testing.T) *echo.Echo {
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
	e := handlers.EchoServer(settings)
	e.Logger.SetLevel(log.ERROR)
	e.Logger.SetHeader("${time_rfc3339} [Glim] ⇨")
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} [REST] ⇨ ${status} ${method} ${uri} ${remote_ip} ${error}\n",
	}))
	e.Logger.Printf("starting REST API in address %s...", settings.Address)
	return e
}

func testCleanUp() {
	removeDatabase()
	removeKV()
}

func removeDatabase() {
	os.Remove("/tmp/test.db")
}

func removeKV() {
	os.RemoveAll("/tmp/kv")
}
