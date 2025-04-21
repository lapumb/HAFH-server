package http

import (
	"github.com/gin-gonic/gin"
	"hafh-server/internal/http/handlers"
	"hafh-server/internal/http/middleware"
)

// Start the HTTP server where an API key is required for all requests.
//
// The server will listen on the specified port and limit the number of requests
// to the specified maximum requests per second.
//
// If the server fails to start, it will panic with an error message.
func StartServer(port string, apiKey string, maxRequestsPerSecond int) {
	if port == "" {
		port = "8080"
	} else if maxRequestsPerSecond <= 0 {
		maxRequestsPerSecond = 5
	} else if apiKey == "" {
		panic("API key is required")
	}

	gin.SetMode(gin.ReleaseMode)

	server := gin.New()
	server.Use(middleware.HttpLogger(), gin.Recovery())

	// Middleware to check for a valid API key and rate limit requests.
	server.Use(middleware.APIKeyAuth(apiKey), middleware.RateLimit(maxRequestsPerSecond))

	// Route definitions:
	server.GET("/api/version", handlers.VersionHandler)

	// Run the server.
	err := server.Run(":" + port)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
