package handlers

import (
	"net/http"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

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

	i := models.GetUserInfo(*u)
	return c.JSON(http.StatusOK, i)
}
