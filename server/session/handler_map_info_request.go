package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"image/color"
)

type MapInfoRequestHandler struct{}

func (*MapInfoRequestHandler) Handle(pk packet.Packet, _ *Session, _ *world.Tx, c Controllable) error {
	if h, ok := c.(interface{ RequestMap(int64) }); ok {
		h.RequestMap(pk.(*packet.MapInfoRequest).MapID)
	}
	return nil
}
func (s *Session) SendMap(id int64, pixels []color.RGBA) {
	if len(pixels) != 128*128 {
		return
	}
	s.writePacket(&packet.ClientBoundMapItemData{MapID: id, Dimension: 0, LockedMap: true, Scale: protocol.Option(byte(0)), MapsIncludedIn: protocol.Option([]int64{id}), Width: protocol.Option(int32(128)), Height: protocol.Option(int32(128)), XOffset: protocol.Option(int32(0)), YOffset: protocol.Option(int32(0)), Pixels: protocol.Option(pixels)})
}
