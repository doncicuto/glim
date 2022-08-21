/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	echoSwagger "github.com/swaggo/echo-swagger"

	handler "github.com/doncicuto/glim/server/api/handlers"
	glimMiddleware "github.com/doncicuto/glim/server/api/middleware"
	"github.com/doncicuto/glim/types"
)

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
func echoServer(settings types.APISettings) *echo.Echo {
	// New Echo framework server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Initialize handler
	blacklist := settings.KV
	h := &handler.Handler{DB: settings.DB, KV: blacklist}

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
	u.GET("", h.FindAllUsers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	u.POST("", h.SaveUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	u.GET("/:uid", h.FindUserByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	u.GET("/:username/uid", h.FindUIDFromUsername, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	u.PUT("/:uid", h.UpdateUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsUpdater)
	u.DELETE("/:uid", h.DeleteUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	u.POST("/:uid/passwd", h.Passwd, glimMiddleware.IsBlacklisted(blacklist))

	g := v1.Group("/groups")
	g.Use(middleware.JWT([]byte(settings.APISecret)))
	g.GET("", h.FindAllGroups, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	g.POST("", h.SaveGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.GET("/:gid", h.FindGroupByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.GET("/:group/gid", h.FindGIDFromGroupName, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	g.PUT("/:gid", h.UpdateGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.DELETE("/:gid", h.DeleteGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.POST("/:gid/members", h.AddGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	g.DELETE("/:gid/members", h.RemoveGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

// Server - TODO command
func Server(wg *sync.WaitGroup, shutdownChannel chan bool, settings types.APISettings) {
	defer wg.Done()

	// Get instance of Echo
	e := echoServer(settings)

	// Set logger level
	e.Logger.SetLevel(log.ERROR)
	e.Logger.SetHeader("${time_rfc3339} [Glim] ⇨")
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} [REST] ⇨ ${status} ${method} ${uri} ${remote_ip} ${error}\n",
	}))
	e.Logger.Printf("starting REST API in address %s...", settings.Address)

	go func() {
		if err := e.StartTLS(settings.Address, settings.TLSCert, settings.TLSKey); err != nil {
			e.Logger.Printf("shutting down REST API server...")
		}
	}()

	// Wait for shutdown signals and gracefully shutdown echo server (10 seconds timeout)
	// Reference: https://echo.labstack.com/cookbook/graceful-shutdown
	<-shutdownChannel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
