/*
Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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
	"os"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	handler "github.com/muultipla/glim/server/api/handlers"
	glimMiddleware "github.com/muultipla/glim/server/api/middleware"
	"github.com/muultipla/glim/server/kv"
)

//Settings - TODO comment
type Settings struct {
	DB      *gorm.DB
	KV      kv.Store
	TLSCert string
	TLSKey  string
}

const apiAddr = ":1323"

//Server - TODO command
func Server(wg *sync.WaitGroup, shutdownChannel chan bool, settings Settings) {
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
			if c.Path() == "/login" || c.Path() == "/login/refresh_token" || c.Path() == "/logout" {
				return true
			}
			return false
		},
	}))

	// Initialize handler
	blacklist := settings.KV
	h := &handler.Handler{DB: settings.DB, KV: blacklist}

	// Routes
	e.POST("/login", h.Login)
	e.POST("/login/refresh_token", h.Refresh)
	e.DELETE("/login/refresh_token", h.Logout)

	e.GET("/users", h.FindAllUsers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	e.POST("/users", h.SaveUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.GET("/users/:uid", h.FindUserByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	e.PUT("/users/:uid", h.UpdateUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.DELETE("/users/:uid", h.DeleteUser, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.POST("/users/:uid/passwd", h.Passwd, glimMiddleware.IsBlacklisted(blacklist))

	e.GET("/groups", h.FindAllGroups, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsReader)
	e.POST("/groups", h.SaveGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.GET("/groups/:gid", h.FindGroupByID, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.PUT("/groups/:gid", h.UpdateGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.DELETE("/groups/:gid", h.DeleteGroup, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)

	e.POST("/groups/:gid/members", h.AddGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)
	e.DELETE("/groups/:gid/members", h.RemoveGroupMembers, glimMiddleware.IsBlacklisted(blacklist), glimMiddleware.IsManager)

	// starting API server....
	e.Logger.Printf("starting REST API in address %s...", addr)

	go func() {
		if err := e.StartTLS(addr, settings.TLSCert, settings.TLSKey); err != nil {
			e.Logger.Printf("shutting down REST API server...")
		}
	}()

	// Wait for shutdown signals and gracefully shutdown echo server (10 seconds tiemout)
	// Reference: https://echo.labstack.com/cookbook/graceful-shutdown
	<-shutdownChannel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
