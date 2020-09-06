package handlers

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/muultipla/glim/models"
)

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
