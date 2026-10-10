package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"time"
)

// PacketProcessor runs exclusively on a session's outgoing writer goroutine.
// Implementations must not perform blocking IO or access world entities.
type PacketProcessor interface {
	Process(packet.Packet, time.Time) []packet.Packet
	Tick(time.Time) []packet.Packet
}
