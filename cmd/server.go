// The main entry point for the server application.
package main

import (
	"hafh-server/internal/http"
	"hafh-server/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val == "1" || val == "true" || val == "yes"
}

func main() {
	debug := getEnvBool("DEBUG", false)

	logger.Init(debug)
	log := logger.Named("main")

	// Note: this will only print to stdout if debug is enabled.
	log.Debug("Debug mode is enabled")

	log.Info("Starting hafh-server...")

	// Initialize the servers.
	go http.StartServer("8080", "dummy", 5)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)

	// Accept SIGINT (Ctrl+C) or SIGTERM (e.g., systemd stop)
	log.Info("Waiting for interrupt signal...")
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Exiting...")
}
