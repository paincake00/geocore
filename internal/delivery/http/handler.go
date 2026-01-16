package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paincake00/geocore/internal/delivery/http/middleware"
	"github.com/paincake00/geocore/internal/usecase"
)

// Pinger интерфейс для проверки соединения с сервисами (БД, Redis).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler структура, объединяющая все HTTP-обработчики.
type Handler struct {
	IncidentService *usecase.IncidentService
	GeoService      *usecase.GeoService
	DBPinger        Pinger
	RedisPinger     Pinger
	APIKey          string
}

// NewHandler создает новый экземпляр HTTP-обработчика.
func NewHandler(is *usecase.IncidentService, gs *usecase.GeoService, db Pinger, rds Pinger, apiKey string) *Handler {
	return &Handler{
		IncidentService: is,
		GeoService:      gs,
		DBPinger:        db,
		RedisPinger:     rds,
		APIKey:          apiKey,
	}
}

// InitRoutes инициализирует роутер Gin и настраивает маршруты API.
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/api/v1/system/health", h.healthCheck)

	v1 := router.Group("/api/v1")
	{
		incidents := v1.Group("/incidents")
		// Применяем middleware авторизации только к группе инцидентов
		incidents.Use(middleware.AuthMiddleware(h.APIKey))
		{
			incidents.POST("", h.createIncident)
			incidents.GET("", h.getIncidents)
			incidents.GET("/stats", h.getStats) // Отдельно от /:id
			incidents.GET("/:id", h.getIncident)
			incidents.PUT("/:id", h.updateIncident)
			incidents.DELETE("/:id", h.deleteIncident)
		}

		location := v1.Group("/location")
		{
			location.POST("/check", h.checkLocation)
		}
	}

	return router
}

// healthCheck проверяет состояние сервиса и зависимостей (PostgreSQL, Redis).
func (h *Handler) healthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	if h.DBPinger != nil {
		if err := h.DBPinger.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "db": err.Error()})
			return
		}
	}
	if h.RedisPinger != nil {
		if err := h.RedisPinger.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "redis": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
