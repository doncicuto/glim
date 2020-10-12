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

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/doncicuto/glim/models"
)

//UpdateGroup - TODO comment
func (h *Handler) UpdateGroup(c echo.Context) error {
	var newGroup = make(map[string]interface{})
	g := new(models.Group)
	body := models.JSONGroupBody{}

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
	err := h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&models.Group{}).Error
	if err != nil {
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
				newGroup["name"] = html.EscapeString(strings.TrimSpace(body.Name))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "group name cannot be duplicated"}
		}
	}

	if body.Description != "" {
		newGroup["description"] = html.EscapeString(strings.TrimSpace(body.Description))
	}

	// New update date
	newGroup["updated_at"] = time.Now()

	// Update group
	err = h.DB.Model(&models.Group{}).Where("id = ?", gid).Updates(newGroup).Error
	if err != nil {
		return err
	}

	// Get updated group
	err = h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&g).Error
	if err != nil {
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
			err = h.DB.Model(&g).Association("Members").Clear().Error
			if err != nil {
				return err
			}
		}

		err = h.AddMembers(g, members)
		if err != nil {
			return err
		}
	}

	// Get updated group
	g = new(models.Group)
	err = h.DB.Preload("Members").Model(&models.Group{}).Where("id = ?", gid).First(&g).Error
	if err != nil {
		return err
	}

	// Return group
	showMembers := true
	return c.JSON(http.StatusOK, models.GetGroupInfo(g, showMembers))
}
