// The main entry point for the server application.
package main

// TODO: make HTTP server configurable (port, API key, rate limit)
// TODO: make cert/key paths for MQTT server configurable

import (
	"context"
	"hafh-server/internal/http"
	"hafh-server/internal/logger"
	"hafh-server/internal/mqtt"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	httpServer, err := http.New("8080", "dummy", 5)
	if err != nil {
		log.Fatal(err)
	}

	// Start the HTTP server.
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	mqttServer, err := mqtt.NewServer(nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Start the MQTT server.
	go func() {
		if err := mqttServer.Start("certs/server.crt", "certs/server.key", "certs/ca.crt", 8883); err != nil {
			log.Fatalf("MQTT server failed: %v", err)
		}
	}()
	log.Info("Servers started successfully!")

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)

	// Accept SIGINT (Ctrl+C) or SIGTERM (e.g., systemd stop)
	log.Info("Waiting for interrupt signal...")
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down HTTP server...")
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Info("Shutting down MQTT server...")
	if err := mqttServer.Shutdown(); err != nil {
		log.Fatalf("MQTT server shutdown error: %v", err)
	}

	log.Info("Exiting...")
}
