package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hafh-server/internal/logger"
	"time"
)

var log *zap.SugaredLogger

// A middleware function that logs HTTP requests.
func HttpLogger() gin.HandlerFunc {
	if log == nil {
		log = logger.Named("http")
	}

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		log.Debug("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
