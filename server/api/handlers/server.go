package handlers

import (
	glimMiddleware "github.com/doncicuto/glim/server/api/middleware"
	"github.com/doncicuto/glim/types"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// EchoServer - TODO command
// @title Glim REST API
// @version 1.0
// @description Glim REST API for login/logout, user and group operations. Users and groups require a Bearer Token (JWT) that you can retrieve using login. Please use the project's README for full information about how you can use this token with Swagger.
// @contact.name Miguel Cabrerizo
// @contact.url https://github.com/doncicuto/glim/issues
// @contact.email support@sologitops.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func EchoServer(settings types.APISettings) *echo.Echo {
	// New Echo framework server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Initialize handler
	blacklist := settings.KV
	h := &Handler{DB: settings.DB, KV: blacklist}

	// Routes
	v1 := e.Group("v1")
	v1.POST("/login", func(c echo.Context) error {
		return h.Login(c, settings)
	})
	v1.POST("/login/refresh_token", func(c echo.Context) error {
		return h.Refresh(c, settings)
	})
	v1.DELETE("/login/refresh_token", func(c echo.Context) error {
		return h.Logout(c, settings)
	})

	u := v1.Group("/users")
	u.Use(middleware.JWT([]byte(settings.APISecret)))
	u.GET("", h.FindAllUsers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader(settings.DB))
	u.POST("", h.SaveUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	u.GET("/:uid", h.FindUserByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader(settings.DB))
	u.GET("/:username/uid", h.FindUIDFromUsername, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader(settings.DB))
	u.PUT("/:uid", h.UpdateUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsUpdater)
	u.DELETE("/:uid", h.DeleteUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	u.POST("/:uid/passwd", h.Passwd, glimMiddleware.IsBlacklisted(blacklist))

	g := v1.Group("/groups")
	g.Use(middleware.JWT([]byte(settings.APISecret)))
	g.GET("", h.FindAllGroups, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader(settings.DB))
	g.POST("", h.SaveGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.GET("/:gid", h.FindGroupByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.GET("/:group/gid", h.FindGIDFromGroupName, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader(settings.DB))
	g.PUT("/:gid", h.UpdateGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.DELETE("/:gid", h.DeleteGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.POST("/:gid/members", h.AddGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.DELETE("/:gid/members", h.RemoveGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}
