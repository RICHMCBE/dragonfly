package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ViewEntityLink shows rider sitting on ridden (linked) or getting off it. Entities this
// session has not spawned are skipped; they get the link with their spawn packet instead.
func (s *Session) ViewEntityLink(rider, ridden world.Entity, linked bool) {
	if s == Nop || rider == nil || ridden == nil {
		return
	}
	riderID, ridID := s.entityRuntimeID(rider), s.entityRuntimeID(ridden)
	if riderID == 0 || ridID == 0 {
		return
	}
	typ := byte(protocol.EntityLinkRemove)
	if linked {
		typ = protocol.EntityLinkRider
	}
	s.writePacket(&packet.SetActorLink{EntityLink: protocol.EntityLink{
		RiddenEntityUniqueID: int64(ridID),
		RiderEntityUniqueID:  int64(riderID),
		Type:                 typ,
		Immediate:            true,
	}})
	s.ViewEntityState(rider)
}

// spawnLinks returns the link of e to the entity it sits on, for its spawn packet, when this
// session already sees that entity.
func (s *Session) spawnLinks(e world.Entity) []protocol.EntityLink {
	r, ok := e.(seatedRider)
	if !ok {
		return nil
	}
	h, riding := r.Mounted()
	if !riding {
		return nil
	}
	s.entityMutex.RLock()
	ridID, known := s.entityRuntimeIDs[h]
	s.entityMutex.RUnlock()
	if !known {
		return nil
	}
	return []protocol.EntityLink{{RiddenEntityUniqueID: int64(ridID), RiderEntityUniqueID: int64(s.entityRuntimeID(e)), Type: protocol.EntityLinkRider, Immediate: true}}
}
