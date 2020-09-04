package api

import (
	"os"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	handler "github.com/muultipla/glim/server/api/handlers"
	glimMiddleware "github.com/muultipla/glim/server/api/middleware"
	"github.com/muultipla/glim/server/kv"
)

const apiAddr = ":1323"

//Server - TODO command
func Server(wg *sync.WaitGroup, database *gorm.DB, blacklist kv.Store) {
	defer wg.Done()

	// New Echo framework server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set logger level
	e.Logger.SetLevel(log.ERROR)
	e.Logger.SetHeader("${time_rfc3339} [Glim] ⇨")
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} [REST] ⇨ ${status} ${method} ${uri} ${remote_ip} ${error}\n",
	}))

	// Get server address
	addr, ok := os.LookupEnv("API_SERVER_ADDRESS")
	if !ok {
		addr = apiAddr
	}

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

	// starting API server....
	e.Logger.Printf("starting REST API in address %s...", addr)
	e.Logger.Fatal(e.Start(addr))
}
