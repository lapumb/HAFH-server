package middleware

import (
	"hafh-server/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

// A middleware function that logs HTTP requests.
func HttpLogger(l *zap.SugaredLogger) gin.HandlerFunc {
	if l == nil {
		log = logger.Named("http")
	} else {
		log = l
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
