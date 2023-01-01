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
	"net/http"

	"github.com/doncicuto/glim/common"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type GlimClaims struct {
	jwtID       float64
	jwtReadonly bool
	jwtManager  bool
}

func getJWTClaims(claims jwt.MapClaims) (GlimClaims, error) {
	var glimClaims GlimClaims

	jwtID, ok := claims["uid"].(float64)
	if !ok {
		return GlimClaims{}, &echo.HTTPError{Code: http.StatusNotAcceptable, Message: common.WrongTokenOrMissingMessage}
	}
	glimClaims.jwtID = jwtID

	jwtReadonly, ok := claims["readonly"].(bool)
	if !ok {
		return GlimClaims{}, &echo.HTTPError{Code: http.StatusNotAcceptable, Message: common.WrongTokenOrMissingMessage}
	}
	glimClaims.jwtReadonly = jwtReadonly

	jwtManager, ok := claims["manager"].(bool)
	if !ok {
		return GlimClaims{}, &echo.HTTPError{Code: http.StatusNotAcceptable, Message: common.WrongTokenOrMissingMessage}
	}
	glimClaims.jwtManager = jwtManager

	return glimClaims, nil
}
