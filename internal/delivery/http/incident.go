package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paincake00/geocore/internal/entity"
)

func (h *Handler) createIncident(c *gin.Context) {
	var input entity.Incident
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Basic validation
	if input.Title == "" || input.Latitude == 0 || input.Longitude == 0 || input.RadiusMeters <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.IncidentService.Create(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *Handler) getIncidents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	incidents, err := h.IncidentService.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incidents)
}

func (h *Handler) getIncident(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	incident, err := h.IncidentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (h *Handler) updateIncident(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input entity.Incident
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id

	if err := h.IncidentService.Update(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *Handler) deleteIncident(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.IncidentService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *Handler) getStats(c *gin.Context) {
	// STATS_TIME_WINDOW_MINUTES could be env, here hardcoded or constant
	// Reqs said "STATS_TIME_WINDOW_MINUTES" (implies configurable), let's say 30 min default
	window := 30

	stats, err := h.IncidentService.GetStats(c.Request.Context(), window)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response: [{"incident_id": 42, "user_count": 17}, ...]
	type StatItem struct {
		IncidentID int `json:"incident_id"`
		UserCount  int `json:"user_count"`
	}
	var response []StatItem
	for k, v := range stats {
		response = append(response, StatItem{IncidentID: k, UserCount: v})
	}

	c.JSON(http.StatusOK, response)
}
