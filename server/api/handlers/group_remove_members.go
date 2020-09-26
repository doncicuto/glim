package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

//RemoveGroupMembers - TODO comment
func (h *Handler) RemoveGroupMembers(c echo.Context) error {
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
			return err
		}

		// Delete association
		err = h.DB.Model(&g).Association("Members").Delete(u).Error
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: fmt.Sprintf("could not remove member from group. Error: %v", err)}
		}
	}

	// Return 204
	return c.NoContent(http.StatusNoContent)
}
