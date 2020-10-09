package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
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

	// Bind
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	// Validate
	if u.Username == nil || *u.Username == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required username"}
	}
	if u.Fullname == nil || *u.Fullname == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required fullname"}
	}
	if u.Password == nil || *u.Password == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required password"}
	}
	if u.Email == nil || *u.Email == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required email"}
	}
	if err := checkmail.ValidateFormat(*u.Email); err != nil {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "invalid email"}
	}

	// Check if user already exists
	err := h.DB.Where("username = ?", *u.Username).First(&u).Error
	if !gorm.IsRecordNotFoundError(err) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "user already exists"}
	}

	// Hash password
	hashedPassword, err := models.Hash(*u.Password)
	if err != nil {
		return err
	}
	*u.Password = string(hashedPassword)

	// Add new user
	err = h.DB.Create(&u).Error
	if err != nil {
		return err
	}

	showMemberOf := true
	i := models.GetUserInfo(*u, showMemberOf)
	return c.JSON(http.StatusOK, i)
}
