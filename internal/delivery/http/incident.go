package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paincake00/geocore/internal/entity"
)

// createIncident обрабатывает запрос на создание нового инцидента.
func (h *Handler) createIncident(c *gin.Context) {
	var input entity.Incident
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Базовая валидация
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

// getIncidents возвращает список инцидентов с пагинацией.
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

// getIncident возвращает инцидент по ID.
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

// updateIncident обновляет данные инцидента.
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

// deleteIncident удаляет инцидент.
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

// getStats возвращает статистику по инцидентам (количество уникальных пользователей в зоне).
func (h *Handler) getStats(c *gin.Context) {
	// STATS_TIME_WINDOW_MINUTES может быть в env, здесь хардкод или константа
	// В ТЗ сказано "STATS_TIME_WINDOW_MINUTES" (подразумевает настройку), возьмем дефолт 30 минут
	window := 30

	stats, err := h.IncidentService.GetStats(c.Request.Context(), window)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Формируем ответ: [{"incident_id": 42, "user_count": 17}, ...]
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
