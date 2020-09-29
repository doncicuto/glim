package handlers

import (
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

// FindGroupByID - TODO comment
func (h *Handler) FindGroupByID(c echo.Context) error {
	var g models.Group
	var err error
	gid := c.Param("gid")

	err = h.DB.Preload("Members").Model(&models.Group{}).Where("id = ?", gid).Take(&g).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		return err
	}

	showMembers := true
	i := models.GetGroupInfo(&g, showMembers)
	return c.JSON(http.StatusOK, i)
}

// FindAllGroups - TODO comment
func (h *Handler) FindAllGroups(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	// Retrieve groups from database
	groups := []models.Group{}
	if err := h.DB.
		Preload("Members").
		Model(&models.Group{}).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&groups).Error; err != nil {
		return err
	}

	if len(groups) == 0 {
		return c.JSON(http.StatusOK, []models.GroupInfo{})
	}

	var allGroups []models.GroupInfo
	showMembers := true
	for _, group := range groups {
		allGroups = append(allGroups, *models.GetGroupInfo(&group, showMembers))
	}

	return c.JSON(http.StatusOK, allGroups)
}
