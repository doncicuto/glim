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

package handlers

import (
	"errors"
	"net/http"

	"github.com/doncicuto/glim/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

//DeleteUser - TODO comment
func (h *Handler) DeleteUser(c echo.Context) error {
	var u models.User
	uid := c.Param("uid")
	err := h.DB.Model(&models.User{}).Where("id = ?", uid).Take(&u).Delete(&u).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
