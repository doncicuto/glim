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

	// Get idparam
	uid := c.Param("uid")

	// User id cannot be empty
	if uid == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required user uid"}
	}

	// Bind
	u := new(models.User)
	if err := c.Bind(u); err != nil {
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
	if u.Username != nil && *u.Username != "" {
		err := h.DB.Model(&models.User{}).Where("name = ? AND id <> ?", u.Username, uid).First(&models.User{}).Error
		if err != nil {
			// Does username exist?
			if gorm.IsRecordNotFoundError(err) {
				newUser["username"] = html.EscapeString(strings.TrimSpace(*u.Username))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "username cannot be duplicated"}
		}

	}

	if u.Fullname != nil && *u.Fullname != "" {
		newUser["fullname"] = html.EscapeString(strings.TrimSpace(*u.Fullname))
	}

	if u.Password != nil && *u.Password != "" {
		password, err := models.Hash(*u.Password)
		if err != nil {
			return err
		}
		newUser["password"] = string(password)
	}
	if u.Email != nil && *u.Email != "" {
		if err := checkmail.ValidateFormat(*u.Email); err != nil {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "invalid email"}
		}
		newUser["email"] = u.Email
	}

	if u.Manager != nil {
		newUser["manager"] = *u.Manager
	}

	if u.Readonly != nil {
		newUser["readonly"] = *u.Readonly
	}

	// New update date
	u.UpdatedAt = time.Now()

	// Update user
	err = h.DB.Model(&u).Where("id = ?", uid).Updates(newUser).Error

	if err != nil {
		return err
	}

	// Get updated user
	err = h.DB.Where("id = ?", uid).First(&u).Error
	if err != nil {
		return err
	}

	// Return user
	return c.JSON(http.StatusOK, models.GetUserInfo(*u))
}
