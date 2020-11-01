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

	"github.com/dgrijalva/jwt-go"
	"github.com/doncicuto/glim/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

//UpdateGroup - TODO comment
func (h *Handler) UpdateGroup(c echo.Context) error {
	var modifiedBy = make(map[string]interface{})
	g := new(models.Group)
	u := new(models.User)
	body := models.JSONGroupBody{}

	// Get username that is updating this group
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	uid, ok := claims["uid"].(float64)
	if !ok {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", uid).First(&u).Error; err != nil {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "wrong user attempting to update group"}
	}

	// Get gid
	gid := c.Param("gid")

	// Group cannot be empty
	if gid == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group gid"}
	}

	// Get request body
	if err := c.Bind(&body); err != nil {
		return err
	}

	// Find group

	if err := h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&models.Group{}).Error; err != nil {
		// Does group exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		return err
	}

	// Validate other fields
	if body.Name != "" {
		err := h.DB.Model(&models.Group{}).Where("name = ? AND id <> ?", g.Name, gid).First(&models.Group{}).Error
		if err != nil {
			// Does group name exist?
			if gorm.IsRecordNotFoundError(err) {
				modifiedBy["name"] = html.EscapeString(strings.TrimSpace(body.Name))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "group name cannot be duplicated"}
		}
	}

	if body.Description != "" {
		modifiedBy["description"] = html.EscapeString(strings.TrimSpace(body.Description))
	}

	// New update date
	modifiedBy["updated_at"] = time.Now()
	modifiedBy["updated_by"] = *u.Username

	// Update group
	if err := h.DB.Model(&models.Group{}).Where("id = ?", gid).Updates(modifiedBy).Error; err != nil {
		return err
	}

	// Get updated group
	if err := h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&g).Error; err != nil {
		// Does group exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		return err
	}

	// Update group members
	if body.Members != "" {
		members := strings.Split(body.Members, ",")

		if body.ReplaceMembers {
			// We are going to replace all group members, so let's clear the associations first
			err := h.DB.Model(&g).Association("Members").Clear().Error
			if err != nil {
				return err
			}
		}

		err := h.AddMembers(g, members)
		if err != nil {
			return err
		}
	}

	// Get updated group
	g = new(models.Group)
	if err := h.DB.Preload("Members").Model(&models.Group{}).Where("id = ?", gid).First(&g).Error; err != nil {
		return err
	}

	// Return group
	showMembers := true
	return c.JSON(http.StatusOK, models.GetGroupInfo(g, showMembers))
}
