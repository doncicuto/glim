package handlers

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

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
