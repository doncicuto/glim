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
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

// AddMembers - TODO comment
func (h *Handler) AddMembers(g *models.Group, members []string) error {
	var err error
	// Update group
	for _, member := range members {
		member = strings.TrimSpace(member)
		// Find user
		u := new(models.User)
		err = h.DB.Model(&models.User{}).Where("username = ?", member).Take(&u).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: fmt.Sprintf("user %s not found", member)}
			}
			return err
		}

		// Append association
		err = h.DB.Model(&g).Association("Members").Append(u).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveGroup - TODO comment
func (h *Handler) SaveGroup(c echo.Context) error {
	g := new(models.Group)
	body := models.JSONGroupBody{}

	// Get request body
	if err := c.Bind(&body); err != nil {
		return err
	}

	// Validate body
	if body.Name == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group name"}
	}

	// Check if group already exists
	err := h.DB.Where("name = ?", body.Name).First(&g).Error
	if !gorm.IsRecordNotFoundError(err) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "group already exists"}
	}

	// Prepare new group
	g.Name = &body.Name
	g.Description = &body.Description

	// Create group
	err = h.DB.Create(&g).Error
	if err != nil {
		return err
	}

	// Add members to group
	if body.Members != "" {
		members := strings.Split(body.Members, ",")
		err = h.AddMembers(g, members)
		if err != nil {
			return err
		}
	}

	// Send group information
	showMembers := true
	i := models.GetGroupInfo(g, showMembers)
	return c.JSON(http.StatusOK, i)
}
