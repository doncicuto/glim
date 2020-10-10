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

package handlers

import (
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

//UpdateUser - TODO comment
func (h *Handler) UpdateUser(c echo.Context) error {
	var newUser = make(map[string]interface{})
	u := new(models.User)

	// Get idparam
	uid := c.Param("uid")

	// User id cannot be empty
	if uid == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required user uid"}
	}

	// Bind
	body := new(models.JSONUserBody)
	if err := c.Bind(body); err != nil {
		return err
	}

	// Find user
	err := h.DB.Where("id = ?", uid).First(&models.User{}).Error
	if err != nil {
		// Does user exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}

	// Validate other fields
	if body.Username != "" {
		err := h.DB.Model(&models.User{}).Where("name = ? AND id <> ?", body.Username, uid).First(&models.User{}).Error
		if err != nil {
			// Does username exist?
			if gorm.IsRecordNotFoundError(err) {
				newUser["username"] = html.EscapeString(strings.TrimSpace(body.Username))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "username cannot be duplicated"}
		}

	}

	if body.Fullname != "" {
		newUser["fullname"] = html.EscapeString(strings.TrimSpace(body.Fullname))
	}

	if body.Email != "" {
		if err := checkmail.ValidateFormat(body.Email); err != nil {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "invalid email"}
		}
		newUser["email"] = body.Email
	}

	if body.Manager != nil {
		newUser["manager"] = *body.Manager
	}

	if body.Readonly != nil {
		newUser["readonly"] = *body.Readonly
	}

	// Update date
	newUser["updated_at"] = time.Now()

	// Update user
	err = h.DB.Model(&models.User{}).Where("id = ?", uid).Updates(newUser).Error
	if err != nil {
		// Does user exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}

	// Get updated user
	err = h.DB.Where("id = ?", uid).First(&u).Error
	if err != nil {
		return err
	}

	// Update group members
	if body.MemberOf != "" {
		members := strings.Split(body.MemberOf, ",")

		if body.ReplaceMembersOf {
			// We are going to replace all user memberof, so let's clear the associations first
			err = h.DB.Model(&u).Association("MemberOf").Clear().Error
			if err != nil {
				return err
			}
		}

		err = h.AddMembersOf(u, members)
		if err != nil {
			return err
		}
	}

	// Get updated user
	err = h.DB.Where("id = ?", uid).First(&u).Error
	if err != nil {
		return err
	}

	// Return user
	showMemberOf := true
	return c.JSON(http.StatusOK, models.GetUserInfo(*u, showMemberOf))
}
