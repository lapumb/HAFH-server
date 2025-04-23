package main

// TODO: make HTTP server configurable (port, API key, rate limit)
// TODO: make cert/key paths for MQTT server configurable
// TODO: made database path configurable

import (
	"context"
	"encoding/json"
	"hafh-server/internal/database"
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

	// Initialize the database.
	db, err := database.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v", err)
		}

		log.Info("Database closed successfully!")
	}()
	log.Info("Database initialized successfully!")

	// TESTING ONLY -- REMOVE
	// Add dummy peripherals
	peripherals := []database.Peripheral{
		{SerialNumber: "TEMP001", Type: database.PeripheralTypeSensor},
		{SerialNumber: "ACT002", Type: database.PeripheralTypeActuator},
		{SerialNumber: "CTRL003", Type: database.PeripheralTypeController},
	}

	for _, p := range peripherals {
		if err := db.AddPeripheral(&p); err != nil {
			log.Errorf("failed to add peripheral %s: %v", p.SerialNumber, err)
		}
	}

	// Add dummy readings
	for i := range 10 {
		r := database.Reading{
			SerialNumber: "TEMP001",
			Timestamp:    time.Now().Add(-time.Duration(i) * time.Minute),
			Data: map[string]any{
				"temperature": 20 + i,
				"humidity":    50 + i,
			},
		}
		if err := db.InsertReading(&r); err != nil {
			log.Errorf("failed to insert reading: %v", err)
		}
	}

	// Print out the readings in JSON format
	readings, err := db.GetLastReadings("TEMP001", 10)
	if err != nil {
		log.Errorf("failed to get readings: %v", err)
	} else {
		jsonBytes, err := json.MarshalIndent(readings, "", "  ")
		if err != nil {
			log.Fatalf("failed to marshal readings to JSON: %v", err)
		}

		log.Info(string(jsonBytes))
	}
	// END TESTING ONLY

	// Initialize the HTTP server.
	httpServer, err := http.New(&http.HttpServerConfig{
		Port:                 "8080",
		ApiKey:               "dummy",
		MaxRequestsPerSecond: 5,
		Db:                   db,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Clean up the HTTP server on exit.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}

		log.Info("HTTP server shutdown successfully!")
	}()

	// Start the MQTT server.
	mqttServer, err := mqtt.New(nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := mqttServer.Start("certs/server.crt", "certs/server.key", "certs/ca.crt", 8883); err != nil {
			log.Fatalf("MQTT server failed: %v", err)
		}
	}()

	// Clean up the MQTT server on exit.
	defer func() {
		if err := mqttServer.Shutdown(); err != nil {
			log.Fatalf("MQTT server shutdown error: %v", err)
		}

		log.Info("MQTT server shutdown successfully!")
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)

	// Accept SIGINT (Ctrl+C) or SIGTERM (e.g., systemd stop).
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Exiting...")
}
