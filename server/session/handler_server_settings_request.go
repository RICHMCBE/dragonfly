package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerSettingsRequestHandler handles the client opening the server settings page.
type ServerSettingsRequestHandler struct{}

// Handle ...
func (h *ServerSettingsRequestHandler) Handle(_ packet.Packet, _ *Session, _ *world.Tx, c Controllable) error {
	c.RequestServerSettings()
	return nil
}
