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

	"github.com/badoux/checkmail"
	"github.com/doncicuto/glim/models"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// AddMembersOf - TODO comment
func (h *Handler) AddMembersOf(u *models.User, memberOf []string) error {
	var err error
	// Update group
	for _, member := range memberOf {
		member = strings.TrimSpace(member)
		// Find group
		g := new(models.Group)
		err = h.DB.Model(&models.Group{}).Where("name = ?", member).Take(&g).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: fmt.Sprintf("group %s not found", member)}
			}
			return err
		}

		// Append association
		err = h.DB.Model(&u).Association("MemberOf").Append(g).Error
		if err != nil {
			return err
		}
	}
	return nil
}

//SaveUser - TODO comment
func (h *Handler) SaveUser(c echo.Context) error {
	u := new(models.User)
	body := models.JSONUserBody{}
	// Bind
	if err := c.Bind(&body); err != nil {
		return err
	}

	// Validate
	if body.Username == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required username"}
	}
	u.Username = &body.Username

	if body.GivenName == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required firstname"}
	}
	u.GivenName = &body.GivenName

	if body.Surname == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required lastname"}
	}
	u.Surname = &body.Surname

	if body.Password == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required password"}
	}

	if body.Email == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required email"}
	}
	if err := checkmail.ValidateFormat(body.Email); err != nil {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "invalid email"}
	}
	u.Email = &body.Email

	if body.Manager != nil {
		u.Manager = body.Manager
	}

	if body.Readonly != nil {
		u.Readonly = body.Readonly
	}

	// Check if user already exists
	err := h.DB.Model(&models.User{}).Where("username = ?", body.Username).First(&u).Error
	if !gorm.IsRecordNotFoundError(err) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "user already exists"}
	}

	// Hash password
	hashedPassword, err := models.Hash(body.Password)
	if err != nil {
		return err
	}
	password := string(hashedPassword)
	u.Password = &password

	// Add new user
	err = h.DB.Model(models.User{}).Create(&u).Error
	if err != nil {
		return err
	}

	// Get new user
	err = h.DB.Where("username = ?", body.Username).First(&u).Error
	if err != nil {
		return err
	}

	// Add group members
	if body.MemberOf != "" {
		members := strings.Split(body.MemberOf, ",")
		err = h.AddMembersOf(u, members)
		if err != nil {
			return err
		}
	}

	showMemberOf := true
	i := models.GetUserInfo(*u, showMemberOf)
	return c.JSON(http.StatusOK, i)
}
