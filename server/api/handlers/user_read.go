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
	"strconv"

	"github.com/doncicuto/glim/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

//FindAllUsers - TODO comment
// @Summary      Find all users
// @Description  Find all users
// @Tags         users
// @Accept       json
// @Produce      json
// @Router       /users [get]
func (h *Handler) FindAllUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	// Retrieve users from database
	users := []models.User{}
	if err := h.DB.
		Preload("MemberOf").
		Model(&models.User{}).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&users).Error; err != nil {
		return err
	}

	if len(users) == 0 {
		return c.JSON(http.StatusOK, []models.UserInfo{})
	}

	var allUsers []models.UserInfo
	showMemberOf := true
	for _, user := range users {
		allUsers = append(allUsers, models.GetUserInfo(user, showMemberOf))
	}

	return c.JSON(http.StatusOK, allUsers)
}

//FindUserByID - TODO comment
// @Summary      Find user by id
// @Description  Find user by id
// @Tags         users
// @Accept       json
// @Produce      json
// @Router       /users/:id [get]
func (h *Handler) FindUserByID(c echo.Context) error {
	var u models.User
	var err error
	uid := c.Param("uid")

	err = h.DB.Preload("MemberOf").Model(&models.User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}

	showMemberOf := true
	i := models.GetUserInfo(u, showMemberOf)
	return c.JSON(http.StatusOK, i)
}
