package http

import (
	"context"
	"errors"
	"fmt"
	"hafh-server/internal/database"
	"hafh-server/internal/http/handlers"
	"hafh-server/internal/http/middleware"
	"hafh-server/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HttpServer represents an HTTP server.
type HttpServer struct {
	internalServer *http.Server
	log            *zap.SugaredLogger
	Db             *database.Database
}

// HttpServerConfig holds the configuration for the HTTP server.
type HttpServerConfig struct {
	Port                 int
	ApiKey               string
	MaxRequestsPerSecond int
	Db                   *database.Database
}

const (
	apiPrefix           = "/api/" + handlers.ApiVersionMajor
	versionEndpoint     = apiPrefix + "/version"
	readingsEndpoint    = apiPrefix + "/readings"
	peripheralsEndpoint = apiPrefix + "/peripherals"
)

// NewServer creates a new [HttpServer] instance with the specified port, API key, and max requests per second (rate limit).
//
// Notes:
//
// - The server will listen on the specified port, defaulting to 8080 if not provided.
//
// - The API key is required for authentication.
//
// - The max requests per second is used to limit the rate of incoming requests, defaulting to 5 if not provided.
func NewServer(config *HttpServerConfig) (*HttpServer, error) {
	// Validate the configuration.
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	port := config.Port
	apiKey := config.ApiKey
	maxRequestsPerSecond := config.MaxRequestsPerSecond
	db := config.Db
	if port == 0 {
		port = 8080
	} else if maxRequestsPerSecond <= 0 {
		maxRequestsPerSecond = 5
	} else if apiKey == "" {
		return nil, errors.New("API key is required")
	} else if db == nil {
		return nil, errors.New("database is required")
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

	handlers.Init(db, log)

	// Route definitions:
	server.GET(versionEndpoint, handlers.GetApiVersion)
	server.GET(peripheralsEndpoint, handlers.GetPeripherals)
	server.POST(peripheralsEndpoint, handlers.PostConfigurePeripheral)
	server.POST(readingsEndpoint, handlers.PostReadings)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server,
	}

	return &HttpServer{
		internalServer: s,
		log:            log,
		Db:             db,
	}, nil
}

// Start starts the HTTP server and listens for incoming requests. **This should be called in a separate goroutine.**
func (s *HttpServer) Start() error {
	s.log.Debugf("HTTP server listening on %s", s.internalServer.Addr)
	if err := s.internalServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.New("failed to start server: " + err.Error())
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server, allowing for any ongoing requests to complete.
func (s *HttpServer) Shutdown(ctx context.Context) error {
	s.log.Debug("Shutting down HTTP server...")
	return s.internalServer.Shutdown(ctx)
}
