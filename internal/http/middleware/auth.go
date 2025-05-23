package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// APIKeyAuth is a middleware function that checks for a valid API key in the request header.
func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}
		c.Next()
	}
}
