package middleware

import (
	"bytes"
	"io"
	"time"

	"hafh-server/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// HttpLogger is a middleware that logs HTTP requests and responses, along with other metadata.
func HttpLogger(l *zap.SugaredLogger) gin.HandlerFunc {
	if l == nil {
		log = logger.Named("http")
	} else {
		log = l
	}

	return func(c *gin.Context) {
		start := time.Now()

		wrappedWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = wrappedWriter

		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore
			requestBody = string(bodyBytes)
		}

		c.Next()

		latency := time.Since(start)
		status := wrappedWriter.Status()
		responseBody := wrappedWriter.body.String()

		log.Debug("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_body", requestBody),
			zap.String("response_body", responseBody),
		)
	}
}
