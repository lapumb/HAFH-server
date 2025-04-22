package http

import (
	"context"
	"errors"
	"hafh-server/internal/http/handlers"
	"hafh-server/internal/http/middleware"
	"hafh-server/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HttpServer struct {
	internalServer *http.Server
	log            *zap.SugaredLogger
}

// Create a new [HttpServer] instance with the specified port, API key, and max requests per second (rate limit).
//
// Notes:
//
// - The server will listen on the specified port, defaulting to 8080 if not provided.
//
// - The API key is required for authentication.
//
// - The max requests per second is used to limit the rate of incoming requests, defaulting to 5 if not provided.
func New(port string, apiKey string, maxRequestsPerSecond int) (*HttpServer, error) {
	if port == "" {
		port = "8080"
	} else if maxRequestsPerSecond <= 0 {
		maxRequestsPerSecond = 5
	} else if apiKey == "" {
		return nil, errors.New("API key is required")
	}

	gin.SetMode(gin.ReleaseMode)

	server := gin.New()

	log := logger.Named("http")
	server.Use(
		middleware.HttpLogger(log),
		middleware.APIKeyAuth(apiKey),
		middleware.RateLimit(maxRequestsPerSecond),
		gin.Recovery(),
	)

	// Route definitions:
	server.GET("/api/version", handlers.VersionHandler)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: server,
	}

	return &HttpServer{
		internalServer: s,
		log:            log,
	}, nil
}

// Start the HTTP server and listen for incoming requests. **This should be called in a separate goroutine.**
func (s *HttpServer) Start() error {
	s.log.Debugf("HTTP server listening on %s", s.internalServer.Addr)
	if err := s.internalServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.New("failed to start server: " + err.Error())
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server, waiting for any ongoing requests to finish.
func (s *HttpServer) Shutdown(ctx context.Context) error {
	s.log.Debug("Shutting down HTTP server...")
	return s.internalServer.Shutdown(ctx)
}
