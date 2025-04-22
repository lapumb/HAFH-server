package mqtt

import (
	"crypto/tls"
	"crypto/x509"
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
}

// NewServer creates a new MQTT server instance.
//
// Note: if onDataReceived is nil, the server will not process incoming MQTT messages.
func NewServer(onDataReceived PublishReceiverFn, onDataReceivedArg any) (*MqttServer, error) {
	log := logger.Named("mqtt")
	s := server.New(nil)
	if s == nil {
		return nil, fmt.Errorf("failed to create MQTT server")
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
		return nil, fmt.Errorf("adding auth hook: %w", err)
	}

	// Hook for logging MQTT events.
	err = s.AddHook(new(LoggingHook), log)
	if err != nil {
		return nil, fmt.Errorf("adding logging hook: %w", err)
	}

	// Hook for processing incoming MQTT messages, if applicable.
	if onDataReceived != nil {
		err = s.AddHook(new(PublishReceiverHook), PublishReceiverConfig{
			log:   log,
			fn:    onDataReceived,
			fnArg: onDataReceivedArg,
		})

		if err != nil {
			return nil, fmt.Errorf("adding publish receiver hook: %w", err)
		}
	}

	return &MqttServer{server: s, log: log}, nil
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
func (s *MqttServer) Start(certPath, keyPath, caPath string, mqttPort int) error {
	tlsConfig, err := loadTLSConfig(certPath, keyPath, caPath)
	if err != nil {
		return err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:        "hafh-mqtt-tls",
		Address:   fmt.Sprintf(":%d", mqttPort),
		TLSConfig: tlsConfig,
	})
	if err := s.server.AddListener(tcp); err != nil {
		return err
	}

	s.log.Debugf("MQTT server listening on :%d (TLS)", mqttPort)

	return s.server.Serve()
}

// Shutdown gracefully shuts down the MQTT server.
func (s *MqttServer) Shutdown() error {
	s.log.Debug("Shutting down MQTT server...")
	return s.server.Close()
}
