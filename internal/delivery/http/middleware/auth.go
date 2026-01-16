package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет наличие и валидность API Key в заголовке X-API-Key.
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Если ключ пустой, считаем что авторизация не требуется (для простоты дебага, но небезопасно для прода)
		if apiKey == "" {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}
