package mqtt

import (
	server "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.uber.org/zap"
)

// LoggingHook is a hook that logs MQTT events such as client connections,
// disconnections, and published messages.
type LoggingHook struct {
	server.HookBase
	log *zap.SugaredLogger
}

// ID returns the ID of the hook.
func (h *LoggingHook) ID() string {
	return "logging-hook"
}

// Provides returns true if the hook provides the specified byte.
func (h *LoggingHook) Provides(b byte) bool {
	switch b {
	case server.OnConnectAuthenticate, server.OnAuthPacket, server.OnConnect, server.OnDisconnect, server.OnPublish:
		return true
	default:
		return false
	}
}

// Init initializes the hook with the provided configuration.
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

// OnConnectAuthenticate logs the client authentication event.
func (h *LoggingHook) OnConnectAuthenticate(cl *server.Client, pk packets.Packet) bool {
	h.log.Debugf("Client %s authenticated: %s", cl.ID, string(pk.Payload))
	return true
}

// OnConnect logs the client connection event.
func (h *LoggingHook) OnConnect(cl *server.Client, pk packets.Packet) error {
	h.log.Debugf("Client connected: %s", cl.ID)
	return nil
}

// OnDisconnect logs the client disconnection event.
func (h *LoggingHook) OnDisconnect(cl *server.Client, err error, expire bool) {
	h.log.Debugf("Client disconnected: %s, error: %v, expired: %v", cl.ID, err, expire)
}

// OnAuthPacket logs the authentication packet event.
func (h *LoggingHook) OnAuthPacket(cl *server.Client, pk packets.Packet) (packets.Packet, error) {
	h.log.Debugf("Client %s sent authentication packet: %s", cl.ID, string(pk.Payload))
	return pk, nil
}

// OnPublish logs the published message event.
func (h *LoggingHook) OnPublish(cl *server.Client, pk packets.Packet) (packets.Packet, error) {
	h.log.Debugf("Client %s sent payload to topic %s: %s", cl.ID, pk.TopicName, string(pk.Payload))
	return pk, nil
}
