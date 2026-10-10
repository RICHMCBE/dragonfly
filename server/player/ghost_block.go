package player

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// SendGhostBlock changes only this player's view. The caller restores both
// layers from the current world after its preview ends.
func (p *Player) SendGhostBlock(pos cube.Pos, b world.Block, layer int) {
	if layer < 0 || layer > 1 {
		return
	}
	p.session().ViewBlockUpdate(pos, b, layer)
}
