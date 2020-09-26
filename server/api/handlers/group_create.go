package handlers

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

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
