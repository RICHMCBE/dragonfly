package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"time"
)

// virtualContainer exists only on the session owner's world goroutine. Fake
// chest blocks are sent to this client alone and never placed in the world.
type virtualContainer struct {
	pos                 []cube.Pos
	world               *world.World
	token               int64
	ready               bool
	close, afterRequest func(*world.Tx)
}

func (s *Session) OpenVirtualContainer(tx *world.Tx, pos cube.Pos, title string, inv *inventory.Inventory, close, afterRequest func(*world.Tx)) error {
	if inv.Size() != 27 && inv.Size() != 54 {
		return fmt.Errorf("virtual chest requires 27 or 54 slots")
	}
	s.closeCurrentContainer(tx, false)
	menu := &virtualContainer{pos: []cube.Pos{pos}, world: tx.World(), token: -time.Now().UnixNano(), close: close, afterRequest: afterRequest}
	if inv.Size() == 54 {
		menu.pos = append(menu.pos, pos.Side(cube.FaceEast))
	}
	s.virtualContainer = menu
	s.nextWindowID()
	s.containerOpened.Store(true)
	s.openedWindow.Store(inv)
	s.openedPos.Store(&pos)
	s.openedContainerID.Store(uint32(protocol.ContainerTypeContainer))
	for i, p := range menu.pos {
		chest := block.NewChest()
		chest.Facing, chest.CustomName = cube.North, title
		s.ViewBlockUpdate(p, chest, 0)
		data := chest.EncodeNBT()
		data["x"], data["y"], data["z"] = int32(p.X()), int32(p.Y()), int32(p.Z())
		if len(menu.pos) == 2 {
			pair := menu.pos[1-i]
			data["pairx"], data["pairz"] = int32(pair.X()), int32(pair.Z())
			data["pairlead"] = byte(1 - i)
		}
		s.writePacket(&packet.BlockActorData{Position: protocol.BlockPos{int32(p.X()), int32(p.Y()), int32(p.Z())}, NBTData: data})
	}
	inv.SlotFunc(func(slot int, _, after item.Stack) {
		if s.virtualContainer == menu && menu.ready {
			s.ViewSlotChange(slot, after)
		}
	})
	// Wait for the client to process the fake block before sending ContainerOpen.
	s.writePacket(&packet.NetworkStackLatency{Timestamp: menu.token, NeedsResponse: true})
	// Some proxies consume latency responses. Keep the acknowledgement fast
	// path, but do not leave a menu permanently pending behind such a proxy.
	time.AfterFunc(500*time.Millisecond, func() {
		menu.world.Do(func(tx *world.Tx) {
			select {
			case <-s.closeBackground:
				return
			default:
			}
			if s.virtualContainer == menu {
				s.openVirtualContainerReady(tx, menu)
			}
		})
	})
	return nil
}

func (s *Session) CloseContainer(tx *world.Tx) { s.closeCurrentContainer(tx, false) }

func (s *Session) closeVirtualContainer(tx *world.Tx, clientRequested bool) bool {
	menu := s.virtualContainer
	if menu == nil {
		return false
	}
	s.virtualContainer = nil
	inv := s.openedWindow.Load()
	inv.SlotFunc(func(int, item.Stack, item.Stack) {})
	s.closeWindow(clientRequested)
	if tx.World() == menu.world {
		for _, pos := range menu.pos {
			s.ViewBlockUpdate(pos, tx.Block(pos), 0)
		}
	}
	if menu.close != nil {
		menu.close(tx)
	}
	return true
}

type virtualContainerLatencyHandler struct{}

func (*virtualContainerLatencyHandler) Handle(pk packet.Packet, s *Session, tx *world.Tx, _ Controllable) error {
	response := pk.(*packet.NetworkStackLatency)
	menu := s.virtualContainer
	if menu == nil || menu.ready || response.NeedsResponse || response.Timestamp != menu.token {
		return nil
	}
	s.openVirtualContainerReady(tx, menu)
	return nil
}

func (s *Session) openVirtualContainerReady(tx *world.Tx, menu *virtualContainer) {
	if s.virtualContainer != menu || menu.ready {
		return
	}
	if menu.world != tx.World() {
		s.closeVirtualContainer(tx, false)
		return
	}
	menu.ready = true
	pos := menu.pos[0]
	s.writePacket(&packet.ContainerOpen{WindowID: byte(s.openedWindowID.Load()), ContainerType: protocol.ContainerTypeContainer,
		ContainerPosition: protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())}, ContainerEntityUniqueID: -1})
	s.sendInv(s.openedWindow.Load(), s.openedWindowID.Load())
}
