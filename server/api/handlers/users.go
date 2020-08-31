package handlers

import (
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

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

//FindAllUsers - TODO comment
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
	for _, user := range users {
		allUsers = append(allUsers, *models.GetUserInfo(user))
	}

	return c.JSON(http.StatusOK, allUsers)
}

//FindUserByID - TODO comment
func (h *Handler) FindUserByID(c echo.Context) error {
	var u models.User
	var err error
	uid := c.Param("uid")

	err = h.DB.Model(&models.User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}

	i := models.GetUserInfo(u)
	return c.JSON(http.StatusOK, i)
}

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

//DeleteUser - TODO comment
func (h *Handler) DeleteUser(c echo.Context) error {
	var u models.User
	uid := c.Param("uid")
	err := h.DB.Model(&models.User{}).Where("id = ?", uid).Take(&u).Delete(&u).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "user not found"}
		}
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
