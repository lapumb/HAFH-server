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

	"go.uber.org/zap"
)

type publishReceiverArg struct {
	log *zap.SugaredLogger
	db  *database.Database
}

func onMqttDataReceived(topic, payload string, arg any) error {
	args, ok := arg.(*publishReceiverArg)
	if !ok {
		panic("invalid argument type")
	}

	// We only care about data published to the "peripherals/readings/#" topic.
	if topic[:len("/peripherals/readings/")] != "/peripherals/readings/" {
		args.log.Debugf("Ignoring topic %s", topic)
		return nil
	}

	// Validate the reading payload.
	reading, err := database.ReadingFromJson([]byte(payload))
	if err != nil {
		return err
	} else if reading == nil || reading.SerialNumber == "" {
		return nil
	}

	// The reading is valid. If the peripheral does not exist, create it.
	peripheral, err := args.db.GetPeripheralBySerial(reading.SerialNumber)
	if err != nil {
		return err
	} else if peripheral == nil {
		args.db.AddPeripheral(&database.Peripheral{
			SerialNumber: reading.SerialNumber,
			Type:         database.PeripheralTypeUnknown,
		})

		args.log.Infof("Added new peripheral: %s", reading.SerialNumber)
	}

	// Insert the reading into the database.
	if err := args.db.InsertReading(reading); err != nil {
		return err
	}

	args.log.Infof("Inserted reading: %s", reading.String())
	return nil
}

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
	httpServer, err := http.New(&http.HttpServerConfig{
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
	forwarder, err := forward.New(&forward.ForwarderConfig{
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

	// Start the MQTT server.
	mqttServer, err := mqtt.New(&mqtt.MqttServerConfig{
		Address:        config.MQTT.Address,
		Port:           config.MQTT.Port,
		CertPath:       config.MQTT.CertPath,
		KeyPath:        config.MQTT.KeyPath,
		CaPath:         config.MQTT.CaPath,
		OnDataReceived: onMqttDataReceived,
		OnDataReceivedArg: &publishReceiverArg{
			log: logger.Named("mqtt::data-received"),
			db:  db,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := mqttServer.Start(); err != nil {
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
