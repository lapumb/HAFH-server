package mqtt

import (
	server "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.uber.org/zap"
)

type LoggingHook struct {
	server.HookBase
	log *zap.SugaredLogger
}

func (h *LoggingHook) ID() string {
	return "logging-hook"
}

func (h *LoggingHook) Provides(b byte) bool {
	switch b {
	case server.OnConnect, server.OnDisconnect, server.OnPublish:
		return true
	default:
		return false
	}
}

func (h *LoggingHook) Init(config any) error {
	if _, ok := config.(*zap.SugaredLogger); !ok && config != nil {
		return server.ErrInvalidConfigType
	}

	h.log = config.(*zap.SugaredLogger)
	if h.log == nil {
		return server.ErrInvalidConfigType
	}

	return nil
}

func (h *LoggingHook) OnConnect(cl *server.Client, pk packets.Packet) error {
	h.log.Debugf("Client connected: %s", cl.ID)
	return nil
}

func (h *LoggingHook) OnDisconnect(cl *server.Client, err error, expire bool) {
	h.log.Debugf("Client disconnected: %s, error: %v, expired: %v", cl.ID, err, expire)
}

func (h *LoggingHook) OnPublish(cl *server.Client, pk packets.Packet) (packets.Packet, error) {
	h.log.Debugf("Client %s sent payload to topic %s: %s", cl.ID, pk.TopicName, string(pk.Payload))

	return pk, nil
}
