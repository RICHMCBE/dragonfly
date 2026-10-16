package session

import (
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"testing"
)

func TestNamedParticlePacketPreservesWorldPosition(t *testing.T) {
	s := &Session{packets: make(chan packet.Packet, 2), closeBackground: make(chan struct{})}
	s.ViewParticle(mgl64.Vec3{10, 65.07, -2}, particle.Named{Identifier: "blf:guardian_roots_15", Dimension: 2})
	p, ok := (<-s.packets).(*packet.SpawnParticleEffect)
	if !ok || p.ParticleName != "blf:guardian_roots_15" || p.EntityUniqueID != -1 || p.Dimension != 2 || p.Position[0] != 10 || p.Position[2] != -2 {
		t.Fatal("custom particle lost identity or absolute position")
	}
	s.ViewParticle(mgl64.Vec3{}, particle.Named{})
	if len(s.packets) != 0 {
		t.Fatal("empty identifier must not be sent")
	}
}
