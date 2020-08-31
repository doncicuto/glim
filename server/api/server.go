package api

import (
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Sqlite3 database
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	handler "github.com/muultipla/glim/server/api/handlers"
	glimMiddleware "github.com/muultipla/glim/server/api/middleware"

	"github.com/muultipla/glim/server/db"
)

//Server - TODO command
func Server() {

	// Get environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	}

	// New Echo framework server
	e := echo.New()

	// Set logger level
	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Logger())

	// Hide banner
	e.HideBanner = true

	// Database
	database, err := db.Initialize()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer database.Close()

	// Key-value store for
	options := badger.DefaultOptions("./server/kvstore/badger")
	// In Windows: To avoid "Value log truncate required to run DB. This might result in data loss" we add the options.Truncate = true
	// Reference: https://discuss.dgraph.io/t/lock-issue-on-windows-on-exposed-api/6316.
	if runtime.GOOS == "windows" {
		options.Truncate = true
	}

	options.Truncate = true
	blacklist, err := badger.Open(options)
	if err != nil {
		log.Fatal(err)
	}
	defer blacklist.Close()

	// JWT tokens will be used for all endpoints but for token requests
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(os.Getenv("API_SECRET")),
		Skipper: func(c echo.Context) bool {
			if c.Path() == "/login" || c.Path() == "/login/refreshToken" || c.Path() == "/logout" {
				return true
			}
			return false
		},
	}))

	// Initialize handler
	h := &handler.Handler{DB: database, KV: blacklist}

	// Routes
	e.POST("/login", h.Login)
	e.POST("/login/refreshToken", h.TokenRefresh)
	e.DELETE("/login/refreshToken", h.Logout)

	// TODO - logout we may receive an expired access token but a valid refresh token so
	// we should blacklist the refresh token

	e.GET("/users", h.FindAllUsers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	e.POST("/users", h.SaveUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.GET("/users/:uid", h.FindUserByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	e.PUT("/users/:uid", h.UpdateUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.DELETE("/users/:uid", h.DeleteUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)

	e.GET("/groups", h.FindAllGroups, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	e.POST("/groups", h.SaveGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.GET("/groups/:gid", h.FindGroupByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.PUT("/groups/:uid", h.UpdateGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.DELETE("/groups/:gid", h.DeleteGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.GET("/groups/:gid/members", h.FindGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)

	// Start our server
	e.Logger.Fatal(e.Start(":1323"))
}
