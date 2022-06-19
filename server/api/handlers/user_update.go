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
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/dgrijalva/jwt-go"
	"github.com/doncicuto/glim/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RemoveMembersOf - TODO comment
func (h *Handler) RemoveMembersOf(u *models.User, memberOf []string) error {
	var err error
	// Update group
	for _, member := range memberOf {
		member = strings.TrimSpace(member)
		// Find group
		g := new(models.Group)
		err = h.DB.Model(&models.Group{}).Where("name = ?", member).Take(&g).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: fmt.Sprintf("group %s not found", member)}
			}
			return err
		}

		// Delete association
		err = h.DB.Model(&u).Association("MemberOf").Delete(g)
		if err != nil {
			return err
		}
	}
	return nil
}

//UpdateUser - TODO comment
func (h *Handler) UpdateUser(c echo.Context) error {
	var updatedUser = make(map[string]interface{})

	u := new(models.User)

	// Get username that is updating this user
	modifiedBy := new(models.User)
	tokenUser := c.Get("user").(*jwt.Token)
	claims := tokenUser.Claims.(jwt.MapClaims)
	tokenUID, ok := claims["uid"].(float64)
	if !ok {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", tokenUID).First(&modifiedBy).Error; err != nil {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "wrong user attempting to update group"}
	}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}

	// Validate other fields
	if body.Username != "" {
		err := h.DB.Model(&models.User{}).Where("name = ? AND id <> ?", body.Username, uid).First(&models.User{}).Error
		if err != nil {
			// Does username exist?
			if errors.Is(err, gorm.ErrRecordNotFound) {
				updatedUser["username"] = html.EscapeString(strings.TrimSpace(body.Username))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "username cannot be duplicated"}
		}
	}

	if body.GivenName != "" {
		updatedUser["fistname"] = html.EscapeString(strings.TrimSpace(body.GivenName))
	}

	if body.Surname != "" {
		updatedUser["lastname"] = html.EscapeString(strings.TrimSpace(body.Surname))
	}

	if body.Email != "" {
		if err := checkmail.ValidateFormat(body.Email); err != nil {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "invalid email"}
		}
		updatedUser["email"] = body.Email
	}

	if body.Manager != nil {
		updatedUser["manager"] = *body.Manager
	}

	if body.Readonly != nil {
		updatedUser["readonly"] = *body.Readonly
	}

	if body.ReplaceMembersOf && body.RemoveMembersOf {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "replace and replace are mutually exclusive"}
	}

	// Update date
	updatedUser["updated_at"] = time.Now()
	updatedUser["updated_by"] = *modifiedBy.Username

	// Update user
	err = h.DB.Model(&models.User{}).Where("id = ?", uid).Updates(updatedUser).Error
	if err != nil {
		// Does user exist?
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
			err = h.DB.Model(&u).Association("MemberOf").Clear()
			if err != nil {
				return err
			}
		}

		if body.RemoveMembersOf {
			err = h.RemoveMembersOf(u, members)
			if err != nil {
				return err
			}
		} else {
			err = h.AddMembersOf(u, members)
			if err != nil {
				return err
			}
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
