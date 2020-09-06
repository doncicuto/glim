package handlers

import (
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

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
