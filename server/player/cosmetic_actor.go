package player

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/go-gl/mathgl/mgl64"
)

func (p *Player) ShowCosmeticActor(identifier string, pos mgl64.Vec3, yaw float64, scale float32, variant, mark int32) uint64 {
	if p.session() == session.Nop {
		return 0
	}
	return p.session().ShowCosmeticActor(identifier, pos, yaw, scale, variant, mark)
}
func (p *Player) ShowCosmeticItem(stack item.Stack, pos mgl64.Vec3) uint64 {
	if p.session() == session.Nop {
		return 0
	}
	return p.session().ShowCosmeticItem(stack, pos)
}
func (p *Player) MoveCosmeticActor(id uint64, pos mgl64.Vec3, yaw float64) {
	if p.session() != session.Nop {
		p.session().MoveCosmeticActor(id, pos, yaw)
	}
}
func (p *Player) HideCosmeticActor(id uint64) {
	if p.session() != session.Nop {
		p.session().HideCosmeticActor(id)
	}
}
