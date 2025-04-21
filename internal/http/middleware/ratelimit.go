package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Middleware to limit the number of requests to 5 per second.
func RateLimit(maxRequestsPerSecond int) gin.HandlerFunc {
	limiter := make(chan time.Time, maxRequestsPerSecond)

	// Fill the limiter.
	for range maxRequestsPerSecond {
		limiter <- time.Now()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for t := range ticker.C {
			select {
			case limiter <- t:
			default:
			}
		}
	}()

	return func(c *gin.Context) {
		select {
		case <-limiter:
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		}
	}
}
