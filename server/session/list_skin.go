package session

import (
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// UpdateListSkin also reaches players in other worlds and the changing player.
// The immutable snapshot is used for players joining after a skin change.
func (s *Session) UpdateListSkin(sk skin.Skin) {
	if s == Nop {
		return
	}
	s.listSkin.Store(&sk)
	encoded := skinToProtocol(sk)
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	for _, other := range sessions.s {
		other.writePacket(&packet.PlayerSkin{UUID: s.ent.UUID(), Skin: encoded})
	}
}
