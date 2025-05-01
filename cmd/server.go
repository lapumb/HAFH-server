package main

import (
	"context"
	"fmt"
	"hafh-server/internal/config"
	"hafh-server/internal/database"
	"hafh-server/internal/http"
	"hafh-server/internal/logger"
	"hafh-server/internal/mqtt"
	forward "hafh-server/internal/ngrok"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const dataTopicPrefix string = "/peripherals/readings/"

func getConfigPath() string {
	if len(os.Args) < 2 {
		return ""
	}

	return os.Args[1]
}

func main() {
	config, err := config.Load(getConfigPath())
	if err != nil && config == nil {
		panic(err)
	}

	logger.Init(config.Debug)
	log := logger.Named("main")

	fmt.Print(`

██╗  ██╗ █████╗ ███████╗██╗  ██╗       /\
██║  ██║██╔══██╗██╔════╝██║  ██║      /  \
███████║███████║█████╗  ███████║     /____\
██╔══██║██╔══██║██╔══╝  ██╔══██║    |      |
██║  ██║██║  ██║██║     ██║  ██║    |  []  |
╚═╝  ╚═╝╚═╝  ╚═╝══╝     ╚═╝  ╚═╝    |______|

 🌐 Welcome to your Home Away from Home 🌐

`)

	// Note: this will only print to stdout if debug is enabled.
	log.Debugf("Using config:\n%s", config.String())

	// Initialize the database.
	db, err := database.New(config.DB.Path)
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

	// Initialize the HTTP server.
	httpServer, err := http.NewServer(&http.HttpServerConfig{
		Port:                 config.HTTP.Port,
		ApiKey:               config.HTTP.APIKey,
		MaxRequestsPerSecond: config.HTTP.MaxRequestsPerSecond,
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

	// Initialize the ngrok forwarder.
	if config.Ngrok.Enabled {
		forwarder, err := forward.NewForwarder(&forward.ForwarderConfig{
			BackendUrl: fmt.Sprintf("localhost:%d", config.HTTP.Port),
			DomainUrl:  config.Ngrok.Domain,
			AuthToken:  config.Ngrok.AuthToken,
			Region:     config.Ngrok.Region,
		})
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			if err := forwarder.Start(context.Background()); err != nil {
				log.Fatalf("Ngrok forwarder failed: %v", err)
			}
		}()
	}

	// Start the MQTT server.
	mqttBroker, err := mqtt.NewBroker(&mqtt.MqttServerConfig{
		Address:         config.MQTT.Address,
		Port:            config.MQTT.Port,
		CertPath:        config.MQTT.CertPath,
		KeyPath:         config.MQTT.KeyPath,
		CaPath:          config.MQTT.CaPath,
		Db:              db,
		DataTopicPrefix: dataTopicPrefix,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := mqttBroker.Start(); err != nil {
			log.Fatalf("Starting MQTT broker failed: %v", err)
		}
	}()

	// Clean up the MQTT broker on exit.
	defer func() {
		if err := mqttBroker.Shutdown(); err != nil {
			log.Fatalf("MQTT broker shutdown error: %v", err)
		}

		log.Info("MQTT broker shutdown successfully!")
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)

	// Accept SIGINT (Ctrl+C) or SIGTERM (e.g., systemd stop).
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Exiting...")
}
