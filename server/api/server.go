package api

import (
	"os"
	"runtime"
	"sync"

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
func Server(wg *sync.WaitGroup) {
	defer wg.Done()

	// Get environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	}

	// New Echo framework server
	e := echo.New()

	// Set logger level
	e.Logger.SetLevel(log.ERROR)
	e.Logger.SetHeader("${time_rfc3339} [Glim] ⇨")
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} [REST] ⇨ ${status} ${method} ${uri} ${remote_ip} ${error}\n",
	}))

	// Hide Echo banner and listening port
	e.HideBanner = true
	e.HidePort = true

	// starting API server....
	e.Logger.Print("starting REST API in port 1323...")

	// Database
	database, err := db.Initialize()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer database.Close()
	e.Logger.Print("REST API connected to database...")

	// Key-value store for
	options := badger.DefaultOptions("./server/kv")
	// In Windows: To avoid "Value log truncate required to run DB. This might result in data loss" we add the options.Truncate = true
	// Reference: https://discuss.dgraph.io/t/lock-issue-on-windows-on-exposed-api/6316.
	if runtime.GOOS == "windows" {
		options.Truncate = true
	}

	// TODO - Enable or disable badger logging
	options.Logger = nil
	blacklist, err := badger.Open(options)
	if err != nil {
		log.Fatal(err)
	}
	defer blacklist.Close()
	e.Logger.Print("REST API connected to key-value store...")

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
	e.Start(":1323")
}
