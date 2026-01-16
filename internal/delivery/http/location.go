package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckLocationInput struct {
	UserID    string  `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func (h *Handler) checkLocation(c *gin.Context) {
	var input CheckLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	matches, err := h.GeoService.CheckLocation(c.Request.Context(), input.UserID, input.Latitude, input.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, matches)
}
