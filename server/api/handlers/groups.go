package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

//UpdateGroup - TODO comment
func (h *Handler) UpdateGroup(c echo.Context) error {
	var newGroup = make(map[string]interface{})

	// Get gid
	gid := c.Param("gid")

	// Group cannot be empty
	if gid == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group gid"}
	}

	// Bind
	g := new(models.Group)
	if err := c.Bind(g); err != nil {
		return err
	}

	// Find group
	err := h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&models.Group{}).Error
	if err != nil {
		// Does group exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		return err
	}

	// Validate other fields
	if g.Name != nil && *g.Name != "" {
		err := h.DB.Model(&models.Group{}).Where("name = ? AND id <> ?", g.Name, gid).First(&models.Group{}).Error
		if err != nil {
			// Does group name exist?
			if gorm.IsRecordNotFoundError(err) {
				newGroup["name"] = html.EscapeString(strings.TrimSpace(*g.Name))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "group name cannot be duplicated"}
		}
	}

	if g.Description != nil && *g.Description != "" {
		newGroup["description"] = html.EscapeString(strings.TrimSpace(*g.Description))
	}

	// New update date
	g.UpdatedAt = time.Now()

	// Update user
	err = h.DB.Model(&g).Where("id = ?", gid).Updates(newGroup).Error

	if err != nil {
		return err
	}

	// Get updated group
	err = h.DB.Model(&g).Where("id = ?", gid).First(&g).Error
	if err != nil {
		return err
	}

	// Return group
	return c.JSON(http.StatusOK, models.GetGroupInfo(g))
}

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

	i := models.GetGroupInfo(&g)
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
	for _, group := range groups {
		allGroups = append(allGroups, *models.GetGroupInfo(&group))
	}

	return c.JSON(http.StatusOK, allGroups)
}

// SaveGroup - TODO comment
func (h *Handler) SaveGroup(c echo.Context) error {
	g := new(models.Group)

	// Get request body
	if err := c.Bind(&g); err != nil {
		return err
	}

	// Validate body
	if g.Name == nil || *g.Name == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group name"}
	}

	// Check if group already exists
	err := h.DB.Where("name = ?", *g.Name).First(&g).Error
	if !gorm.IsRecordNotFoundError(err) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "group already exists"}
	}

	// Create group
	err = h.DB.Create(&g).Error
	if err != nil {
		return err
	}

	// Send group information
	i := models.GetGroupInfo(g)
	return c.JSON(http.StatusOK, i)
}

//DeleteGroup - TODO comment
func (h *Handler) DeleteGroup(c echo.Context) error {
	var g models.Group
	gid := c.Param("gid")
	err := h.DB.Model(&g).Where("id = ?", gid).Take(&g).Delete(&g).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

//AddGroupMembers - TODO comment
func (h *Handler) AddGroupMembers(c echo.Context) error {
	// Get gid
	gid := c.Param("gid")

	// Group cannot be empty
	if gid == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group gid"}
	}

	// Bind
	m := new(models.GroupMembers)
	if err := c.Bind(m); err != nil {
		return err
	}
	members := strings.Split(m.Members, ",")

	// Find group
	g := new(models.Group)
	err := h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&g).Error
	if err != nil {
		// Does group exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		fmt.Println("1")
		return err
	}

	// Update group
	for _, member := range members {

		// Find user
		u := new(models.User)
		err = h.DB.Model(&models.User{}).Where("username = ?", member).Take(&u).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: fmt.Sprintf("user %s not found", member)}
			}
			fmt.Println("2")
			return err
		}

		// Append association
		err = h.DB.Model(&g).Association("Members").Append(u).Error
		if err != nil {
			fmt.Println("3")
			return err
		}
	}

	// Get updated group
	err = h.DB.Model(&g).Where("id = ?", gid).First(&g).Error
	if err != nil {
		return err
	}

	// Return group
	return c.JSON(http.StatusOK, models.GetGroupInfo(g))
}
