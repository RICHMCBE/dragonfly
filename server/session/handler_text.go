package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TextHandler handles the Text packet.
type TextHandler struct{}

// Handle ...
func (TextHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.Text)

	if pk.TextType != packet.TextTypeChat {
		return fmt.Errorf("TextType should always be Chat (%v), but got %v", packet.TextTypeChat, pk.TextType)
	}
	// SourceName is presentation metadata. A client may echo the decorated
	// PlayerList username (for example an admin/guide badge). Never use this
	// untrusted field for attribution: c belongs to this authenticated session
	// and Chat uses that player's server-owned identity.
	if pk.XUID != s.conn.IdentityData().XUID {
		return fmt.Errorf("XUID must be equal to player's XUID")
	}
	c.Chat(pk.Message)
	return nil
}
