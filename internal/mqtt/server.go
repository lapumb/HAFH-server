package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
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
	Address           string
	Port              int
	CertPath          string
	KeyPath           string
	CaPath            string
	OnDataReceived    PublishReceiverFn
	OnDataReceivedArg any
}

// New creates a new MQTT server instance.
//
// Note: if onDataReceived is nil, the server will not process incoming MQTT messages.
func New(config *MqttServerConfig) (*MqttServer, error) {
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
	if config.OnDataReceived != nil {
		err = s.AddHook(new(PublishReceiverHook), PublishReceiverConfig{
			log:   log,
			fn:    config.OnDataReceived,
			fnArg: config.OnDataReceivedArg,
		})

		if err != nil {
			return nil, errors.New("failed to add publish receiver hook: " + err.Error())
		}
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
