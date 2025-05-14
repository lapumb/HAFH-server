package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"hafh-server/internal/database"
	"hafh-server/internal/logger"
	"log/slog"
	"os"

	server "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"go.uber.org/zap"
)

// MqttServer represents an MQTT server.
type MqttServer struct {
	server *server.Server
	log    *zap.SugaredLogger
	config MqttServerConfig
}

// MqttServerConfig holds the configuration for the MQTT server.
type MqttServerConfig struct {
	Address         string
	Port            int
	CertPath        string
	KeyPath         string
	CaPath          string
	Db              *database.Database
	DataTopicPrefix string
}

type publishReceiverArg struct {
	log             *zap.SugaredLogger
	db              *database.Database
	dataTopicPrefix string
}

// NewBroker creates a new MQTT broker (server) instance.
//
// Note: if onDataReceived is nil, the broker will not directly process incoming MQTT messages.
func NewBroker(config *MqttServerConfig) (*MqttServer, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	} else if config.CertPath == "" || config.KeyPath == "" || config.CaPath == "" {
		return nil, errors.New("certPath, keyPath, and caPath cannot be empty")
	}

	log := logger.Named("mqtt")
	s := server.New(nil)
	if s == nil {
		return nil, errors.New("failed to create MQTT server")
	}

	// Only log errors and above.
	level := new(slog.LevelVar)
	s.Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	level.Set(slog.LevelError)

	// Do not require username and password for authentication (TLS will be used).
	err := s.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		return nil, errors.New("failed to add auth hook: " + err.Error())
	}

	// Hook for logging MQTT events.
	err = s.AddHook(new(LoggingHook), log)
	if err != nil {
		return nil, errors.New("failed to add logging hook: " + err.Error())
	}

	// Hook for processing incoming MQTT messages, if applicable.
	if config.DataTopicPrefix != "" && config.Db != nil {
		err = s.AddHook(new(PublishReceiverHook), PublishReceiverConfig{
			log:   log,
			fn:    onMqttDataReceived,
			fnArg: &publishReceiverArg{log: log, db: config.Db, dataTopicPrefix: config.DataTopicPrefix},
		})

		if err != nil {
			return nil, errors.New("failed to add publish receiver hook: " + err.Error())
		}
	} else {
		log.Debug("Skipping publish receiver hook as no data topic prefix or database is provided")
	}

	tlsConfig, err := loadTLSConfig(config.CertPath, config.KeyPath, config.CaPath)
	if err != nil {
		return nil, errors.New("failed to load TLS config: " + err.Error())
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:        "hafh-mqtt-tls",
		Address:   fmt.Sprintf("%s:%d", config.Address, config.Port),
		TLSConfig: tlsConfig,
	})
	if err := s.AddListener(tcp); err != nil {
		return nil, errors.New("failed to add TCP listener: " + err.Error())
	}

	internal := MqttServer{server: s, log: log, config: *config}
	if internal.config.Port == 0 {
		internal.config.Port = 8883
	}

	return &internal, nil
}

// Start starts the MQTT server (TLS) and listens for incoming connections on the specified port.
func (s *MqttServer) Start() error {
	s.log.Debugf("MQTT server listening on %s:%d (TLS)", s.config.Address, s.config.Port)
	return s.server.Serve()
}

// Shutdown gracefully shuts down the MQTT server.
func (s *MqttServer) Shutdown() error {
	s.log.Debug("Shutting down MQTT server...")
	return s.server.Close()
}

func loadTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("reading CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("appending CA cert")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("loading cert/key: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func onMqttDataReceived(topic, payload string, arg any) error {
	args, ok := arg.(*publishReceiverArg)
	if !ok {
		panic("invalid argument type")
	}

	// We only care about data published to the specified topic prefix.
	if len(topic) < len(args.dataTopicPrefix) || topic[:len(args.dataTopicPrefix)] != args.dataTopicPrefix {
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
