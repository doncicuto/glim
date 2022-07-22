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
	"net/http"
	"strings"

	"github.com/doncicuto/glim/models"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// AddMembers - TODO comment
func (h *Handler) AddMembers(g *models.Group, members []string) error {
	var err error
	// Update group
	for _, member := range members {
		member = strings.TrimSpace(member)
		// Find user
		createdBy := new(models.User)
		err = h.DB.Model(&models.User{}).Where("username = ?", member).Take(&createdBy).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: fmt.Sprintf("user %s not found", member)}
			}
			return err
		}

		// Append association
		err = h.DB.Model(&g).Association("Members").Append(createdBy)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveGroup - TODO comment
// @Summary      Create a new group
// @Description  Create a new group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        group body models.JSONGroupBody  true  "Group body. Name is required. The members property expect a comma-separated list of usernames e.g 'bob,sally'. The replace property is not used in this command."
// @Success      200  {object}  models.GroupInfo
// @Failure			 400  {object} types.ErrorResponse
// @Failure			 401  {object} types.ErrorResponse
// @Failure 	   404  {object} types.ErrorResponse
// @Failure 	   406  {object} types.ErrorResponse
// @Failure 	   500  {object} types.ErrorResponse
// @Router       /groups [post]
// @Security 		 Bearer
func (h *Handler) SaveGroup(c echo.Context) error {
	g := new(models.Group)
	createdBy := new(models.User)
	body := models.JSONGroupBody{}

	// Get username that is updating this group
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	uid, ok := claims["uid"].(float64)
	if !ok {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "wrong token or missing info in token claims"}
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", uint(uid)).First(&createdBy).Error; err != nil {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "wrong user attempting to update group"}
	}

	// Get request body
	if err := c.Bind(&body); err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	// Validate body
	if body.Name == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group name"}
	}

	// Check if group already exists
	err := h.DB.Where("name = ?", body.Name).First(&g).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "group already exists"}
	}

	// Prepare new UUID
	groupUUID := uuid.New().String()
	g.UUID = &groupUUID

	// Prepare new group
	g.Name = &body.Name
	g.Description = &body.Description

	// Created by
	g.CreatedBy = createdBy.Username
	g.UpdatedBy = createdBy.Username

	// Create group
	err = h.DB.Create(&g).Error
	if err != nil {
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	// Add members to group
	if body.Members != "" {
		members := strings.Split(body.Members, ",")
		err = h.AddMembers(g, members)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()}
		}
	}

	// Send group information
	showMembers := true
	i := models.GetGroupInfo(g, showMembers)
	return c.JSON(http.StatusOK, i)
}
