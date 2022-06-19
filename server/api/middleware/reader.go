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

package middleware

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

//IsReader - TODO comment
func IsReader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		uid := c.Param("uid")
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		jwtID, ok := claims["uid"].(float64)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}
		jwtReadonly, ok := claims["readonly"].(bool)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}
		jwtManager, ok := claims["manager"].(bool)
		if !ok {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
		}
		if jwtManager || !jwtReadonly || (jwtReadonly && uid == fmt.Sprintf("%f", jwtID)) {
			return next(c)
		}
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "user has no proper permissions"}
	}
}
