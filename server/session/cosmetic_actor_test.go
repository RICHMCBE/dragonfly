package session

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"testing"
)

func TestCosmeticActorIdentifierAnnounced(t *testing.T) {
	w := world.New()
	defer w.Close()
	s := &Session{conf: Config{CosmeticActorIdentifiers: []string{"blf:dotori_draw_chest"}}, packets: make(chan packet.Packet, 10), closeBackground: make(chan struct{})}
	s.sendAvailableEntities(w)
	p := (<-s.packets).(*packet.AvailableActorIdentifiers)
	var data struct {
		IDs []actorIdentifier `nbt:"idlist"`
	}
	if err := nbt.Unmarshal(p.SerialisedEntityIdentifiers, &data); err != nil {
		t.Fatal(err)
	}
	for _, id := range data.IDs {
		if id.ID == "blf:dotori_draw_chest" {
			return
		}
	}
	t.Fatal("client cannot resolve cosmetic actor")
}

func TestCosmeticActorsArePrivateAndNotInteractive(t *testing.T) {
	owner := &Session{br: world.DefaultBlockRegistry, currentEntityRuntimeID: 41, packets: make(chan packet.Packet, 20), closeBackground: make(chan struct{})}
	observer := &Session{packets: make(chan packet.Packet, 20), closeBackground: make(chan struct{})}
	pos := mgl64.Vec3{1, 65, 2}
	chest := owner.ShowCosmeticActor("blf:dotori_draw_chest", pos, 180, .8, 1, 1)
	display := owner.ShowCosmeticItem(item.NewStack(item.Diamond{}, 2), pos)
	if chest != 42 || display != 43 {
		t.Fatalf("runtime IDs collide: %d %d", chest, display)
	}
	if len(owner.entities) != 0 || len(observer.packets) != 0 {
		t.Fatal("cosmetics registered as real entities or broadcast")
	}
	add := (<-owner.packets).(*packet.AddActor)
	if add.EntityMetadata[protocol.EntityDataKeyVariant] != int32(1) || add.EntityMetadata[protocol.EntityDataKeyMarkVariant] != int32(1) {
		t.Fatal("quick/pity flags lost")
	}
	if add.EntityMetadata[protocol.EntityDataKeyScale] != float32(.8) {
		t.Fatal("scale lost")
	}
	if p := (<-owner.packets).(*packet.AddItemActor); p.EntityRuntimeID != display || p.Item.Stack.Count != 2 {
		t.Fatal("display item lost")
	}
	owner.MoveCosmeticActor(display, pos, 90)
	if p := (<-owner.packets).(*packet.MoveActorAbsolute); p.EntityRuntimeID != display {
		t.Fatal("wrong actor moved")
	}
	owner.HideCosmeticActor(chest)
	owner.HideCosmeticActor(display)
	if len(owner.cosmeticActors) != 0 {
		t.Fatal("actors leaked")
	}
	<-owner.packets
	<-owner.packets
	owner.MoveCosmeticActor(display, pos, 0)
	owner.HideCosmeticActor(display)
	owner.HideCosmeticActor(1)
	if len(owner.packets) != 0 {
		t.Fatal("removed or real actor changed")
	}
}
