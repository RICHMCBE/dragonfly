package particle

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Named spawns an effect defined in the client's active resource pack.
// It is sent only to viewers of the world transaction adding the particle.
type Named struct {
	Identifier string
	Dimension  byte
}

func (Named) Spawn(*world.World, mgl64.Vec3) {}
