package handlers

import (
	"os"

	"github.com/doncicuto/glim/server/db"
	"github.com/doncicuto/glim/server/kv/badgerdb"
	"github.com/doncicuto/glim/types"
	"gorm.io/gorm"
)

func newTestDatabase() (*gorm.DB, error) {
	var dbInit = types.DBInit{
		AdminPasswd:   "test",
		SearchPasswd:  "test",
		Users:         "saul,kim,mike",
		DefaultPasswd: "test",
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
		Address:            "",
		APISecret:          "secret",
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 259200,
		MaxDaysWoRelogin:   7,
	}
}

func removeDatabase() {
	os.Remove("/tmp/test.db")
}

func removeKV() {
	os.RemoveAll("/tmp/kv")
}
