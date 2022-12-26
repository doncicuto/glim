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
	"time"

	"github.com/doncicuto/glim/common"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// Logout - TODO comment
// @Summary      Delete authentication tokens
// @Description  Log out from the API
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        tokens  body common.Tokens  true  "Access and Refresh JWT tokens"
// @Success      204
// @Failure			 400  {object} common.ErrorResponse
// @Failure			 401  {object} common.ErrorResponse
// @Failure 	   500  {object} common.ErrorResponse
// @Router       /login/refresh_token [delete]
func (h *Handler) Logout(c echo.Context, settings common.APISettings) error {

	// Get refresh token from body
	tokens := new(common.Tokens)
	if err := c.Bind(tokens); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: common.CouldNotParseTokenMessage}
	}

	// Get refresh token claims
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(tokens.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(settings.APISecret), nil
	})

	// Extract access token jti
	ajti, ok := claims["ajti"].(string)
	if !ok {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: common.UseRefreshTokenMessage}
	}

	if token.Valid && err == nil {

		// Extract jti
		jti, ok := claims["jti"].(string)
		if !ok {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: common.JTINotFoundInTokenMessage}
		}

		// Check if token has been blacklisted
		val, found, err := h.KV.Get(jti)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: common.CouldNotGetStoredTokenMessage}
		}
		if found {
			// blacklisted item?
			if string(val) == "true" {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: common.TokenNoLongerValidMessage}
			}
		}

		// Blacklist refresh token
		err = h.KV.Set(jti, "true", time.Second*3600)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: common.CouldNotGetStoredTokenMessage}
		}

		// Blacklist access token
		err = h.KV.Set(ajti, "true", time.Second*3600)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "could not store access token info"}
		}
	}

	return c.NoContent(http.StatusNoContent)
}
