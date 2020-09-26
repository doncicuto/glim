package handlers

import (
	"html"
	"net/http"
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
