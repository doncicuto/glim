/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@sologitops.com>

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

package handlers

import (
	"fmt"
	"net/http"

	"github.com/doncicuto/glim/common"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// IsUpdater - TODO comment
func IsUpdater(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		uid := c.Param("uid")
		if c.Get("user") == nil {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: common.WrongTokenOrMissingMessage}
		}
		user := c.Get("user").(*jwt.Token)
		claims, err := getJWTClaims(user.Claims.(jwt.MapClaims))
		if err != nil {
			return err
		}

		if !claims.jwtReadonly && (claims.jwtManager || uid == fmt.Sprintf("%d", uint(claims.jwtID))) {
			return next(c)
		}
		return &echo.HTTPError{Code: http.StatusForbidden, Message: common.UserHasNoProperPermissionsMessage}
	}
}
