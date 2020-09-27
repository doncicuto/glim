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
	g := new(models.Group)
	body := models.JSONGroupBody{}

	// Get gid
	gid := c.Param("gid")

	// Group cannot be empty
	if gid == "" {
		return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "required group gid"}
	}

	// Get request body
	if err := c.Bind(&body); err != nil {
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
	if body.Name != "" {
		err := h.DB.Model(&models.Group{}).Where("name = ? AND id <> ?", g.Name, gid).First(&models.Group{}).Error
		if err != nil {
			// Does group name exist?
			if gorm.IsRecordNotFoundError(err) {
				newGroup["name"] = html.EscapeString(strings.TrimSpace(body.Name))
			}
		} else {
			return &echo.HTTPError{Code: http.StatusNotAcceptable, Message: "group name cannot be duplicated"}
		}
	}

	if body.Description != "" {
		newGroup["description"] = html.EscapeString(strings.TrimSpace(body.Description))
	}

	// New update date
	g.UpdatedAt = time.Now()

	// Update group
	err = h.DB.Model(&models.Group{}).Where("id = ?", gid).Updates(newGroup).Error
	if err != nil {
		return err
	}

	// Get updated group
	err = h.DB.Model(&models.Group{}).Where("id = ?", gid).First(&g).Error
	if err != nil {
		// Does group exist?
		if gorm.IsRecordNotFoundError(err) {
			return &echo.HTTPError{Code: http.StatusNotFound, Message: "group not found"}
		}
		return err
	}

	// Update group members
	if body.Members != "" {
		members := strings.Split(body.Members, ",")

		if body.ReplaceMembers {
			// We are going to replace all group members, so let's clear the associations first
			err = h.DB.Model(&g).Association("Members").Clear().Error
			if err != nil {
				return err
			}
		}

		err = h.AddMembers(g, members)
		if err != nil {
			return err
		}
	}

	// Get updated group
	g = new(models.Group)
	err = h.DB.Preload("Members").Model(&models.Group{}).Where("id = ?", gid).First(&g).Error
	if err != nil {
		return err
	}

	// Return group
	return c.JSON(http.StatusOK, models.GetGroupInfo(g))
}
