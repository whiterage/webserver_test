package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// middleware for checking API key
func APIKeyAuth(validAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiKey string

		// проверяем заголовок
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				apiKey = parts[1]
			}
		}

		// проверяем X-API-Key
		if apiKey == "" {
			apiKey = c.GetHeader("X-API-Key")
		}

		// Проверяем ключ
		if apiKey == "" || apiKey != validAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		// Ключ валиден, продолжаем
		c.Next()
	}
}
