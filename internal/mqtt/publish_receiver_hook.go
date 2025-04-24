package mqtt

import (
	"bytes"

	server "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.uber.org/zap"
)

// PublishReceiverFn is a function type that processes incoming MQTT messages.
type PublishReceiverFn func(topic, payload string, arg any) error

// PublishReceiverConfig holds the configuration for the PublishReceiverHook.
type PublishReceiverConfig struct {
	log   *zap.SugaredLogger
	fn    PublishReceiverFn
	fnArg any
}

// PublishReceiverHook is a hook that processes incoming MQTT messages
// for internal processing.
type PublishReceiverHook struct {
	server.HookBase
	config PublishReceiverConfig
}

// ID returns the ID of the hook.
func (h *PublishReceiverHook) ID() string {
	return "publish-receiver-hook"
}

// Provides returns true if the hook provides the specified byte.
func (h *PublishReceiverHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		server.OnPublished,
	}, []byte{b})
}

// Init initializes the hook with the provided configuration.
func (h *PublishReceiverHook) Init(config any) error {
	if cfg, ok := config.(PublishReceiverConfig); ok {
		h.config = cfg
		return nil
	} else if cfg, ok := config.(*PublishReceiverConfig); ok {
		h.config = *cfg
		return nil
	}

	return server.ErrInvalidConfigType
}

// OnPublish processes incoming MQTT messages and calls the configured function.
func (h *PublishReceiverHook) OnPublished(cl *server.Client, pk packets.Packet) {
	if h.config.fn == nil {
		return
	}

	if err := h.config.fn(pk.TopicName, string(pk.Payload), h.config.fnArg); err != nil {
		h.config.log.Errorf("Failed to process MQTT message: %v", err)
		return
	}
}
