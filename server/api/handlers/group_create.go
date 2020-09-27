package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

// AddMembers - TODO comment
func (h *Handler) AddMembers(g *models.Group, members []string) error {
	var err error

	// Update group
	for _, member := range members {
		member = strings.TrimSpace(member)
		// Find user
		u := new(models.User)
		err = h.DB.Model(&models.User{}).Where("username = ?", member).Take(&u).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return &echo.HTTPError{Code: http.StatusNotFound, Message: fmt.Sprintf("user %s not found", member)}
			}
			return err
		}

		// Append association
		err = h.DB.Model(&g).Association("Members").Append(u).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveGroup - TODO comment
func (h *Handler) SaveGroup(c echo.Context) error {
	g := new(models.Group)
	body := models.NewGroup{}

	// Get request body
	if err := c.Bind(&body); err != nil {
		return err
	}

	// Validate body
	if body.Name == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group name"}
	}

	// Check if group already exists
	err := h.DB.Where("name = ?", body.Name).First(&g).Error
	if !gorm.IsRecordNotFoundError(err) {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "group already exists"}
	}

	// Prepare new group
	g.Name = &body.Name
	g.Description = &body.Description

	// Create group
	err = h.DB.Create(&g).Error
	if err != nil {
		return err
	}

	// Add members to group
	members := strings.Split(body.Members, ",")
	err = h.AddMembers(g, members)
	if err != nil {
		return err
	}

	// Send group information
	i := models.GetGroupInfo(g)
	return c.JSON(http.StatusOK, i)
}
