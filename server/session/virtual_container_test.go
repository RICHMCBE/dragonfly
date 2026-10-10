package session

import (
	"context"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"testing"
	"time"
)

func TestVirtualChestOpensWithoutLatencyResponse(t *testing.T) {
	w := world.New()
	defer w.Close()
	s := &Session{br: world.DefaultBlockRegistry, packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
	if err := w.Do(func(tx *world.Tx) {
		if err := s.OpenVirtualContainer(tx, cube.Pos{0, 64, 0}, "fallback", inventory.New(27, nil), nil, nil); err != nil {
			t.Fatal(err)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case p := <-s.packets:
			if _, ok := p.(*packet.ContainerOpen); ok {
				return
			}
		case <-timer.C:
			t.Fatal("menu remained pending without latency response")
		}
	}
}

func TestVirtualChestAcknowledgementAndClose(t *testing.T) {
	w := world.New()
	defer w.Close()
	err := w.Do(func(tx *world.Tx) {
		for _, size := range []int{27, 54} {
			s := &Session{br: world.DefaultBlockRegistry, packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
			pos := cube.Pos{0, 64, 0}
			tx.SetBlock(pos, block.Stone{}, nil)
			inv := inventory.New(size, nil)
			_ = inv.SetItem(10, item.NewStack(item.Diamond{}, 3))
			closed := 0
			if err := s.OpenVirtualContainer(tx, pos, "§c§1§btitle", inv, func(*world.Tx) { closed++ }, nil); err != nil {
				t.Fatal(err)
			}
			if _, ok := tx.Block(pos).(block.Stone); !ok {
				t.Fatal("fake chest changed world")
			}
			if _, ok := s.invByID(protocol.ContainerLevelEntity, tx); ok {
				t.Fatal("inventory accessible before acknowledgement")
			}
			h := &virtualContainerLatencyHandler{}
			_ = h.Handle(&packet.NetworkStackLatency{Timestamp: s.virtualContainer.token + 1}, s, tx, nil)
			if s.virtualContainer.ready {
				t.Fatal("stale acknowledgement opened menu")
			}
			_ = h.Handle(&packet.NetworkStackLatency{Timestamp: s.virtualContainer.token}, s, tx, nil)
			if i, ok := s.invByID(protocol.ContainerLevelEntity, tx); !ok || i != inv {
				t.Fatal("menu not accessible")
			}
			if _, ok := s.invByID(protocol.ContainerFurnaceIngredient, tx); ok {
				t.Fatal("fake container exposed unrelated block inventory")
			}
			opens := 0
			for len(s.packets) > 0 {
				if _, ok := (<-s.packets).(*packet.ContainerOpen); ok {
					opens++
				}
			}
			if opens != 1 {
				t.Fatal("expected exactly one open", opens)
			}
			s.CloseContainer(tx)
			s.CloseContainer(tx)
			if closed != 1 || s.virtualContainer != nil || s.containerOpened.Load() {
				t.Fatal("close must be idempotent")
			}
		}
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestVirtualChestFallbackDoesNotReopenClosedMenu(t *testing.T) {
	w := world.New()
	defer w.Close()
	s := &Session{br: world.DefaultBlockRegistry, packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
	if err := w.Do(func(tx *world.Tx) {
		if err := s.OpenVirtualContainer(tx, cube.Pos{0, 64, 0}, "closed", inventory.New(27, nil), nil, nil); err != nil {
			t.Fatal(err)
		}
		s.CloseContainer(tx)
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := w.DoAfter(time.Second, func(*world.Tx) {}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	for len(s.packets) > 0 {
		if _, ok := (<-s.packets).(*packet.ContainerOpen); ok {
			t.Fatal("fallback reopened closed menu")
		}
	}
}

func TestRequestRollbackKeepsFirstSlotSnapshot(t *testing.T) {
	w := world.New()
	defer w.Close()
	err := w.Do(func(tx *world.Tx) {
		s := &Session{br: world.DefaultBlockRegistry, packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{}), inv: inventory.New(36, nil)}
		initial := item.NewStack(item.Diamond{}, 9)
		_ = s.inv.SetItem(0, initial)
		h := &ItemStackRequestHandler{currentRequest: 1, changes: map[byte]map[byte]changeInfo{}, responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{}}
		slot := protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: protocol.ContainerInventory}, Slot: 0}
		h.setItemInSlot(slot, initial.Grow(-1), s, tx)
		h.setItemInSlot(slot, initial.Grow(-3), s, tx)
		h.reject(1, s, tx)
		got, _ := s.inv.Item(0)
		if !got.Equal(initial) {
			t.Fatal("rejected multi-action request did not restore original stack", got)
		}
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
