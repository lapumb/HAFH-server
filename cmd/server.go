// The main entry point for the server application.
package main

import (
	"hafh-server/internal/logger"
	"os"
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

    log.Info("Starting hafh-server...")

	// Initialize the server

	log.Info(("Stopping hafh-server..."))
}
