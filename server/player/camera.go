package player

import (
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func (p *Player) SetFreeCamera(position, target mgl64.Vec3) {
	if p.session() != session.Nop {
		p.session().SetFreeCamera(position, target)
	}
}
func (p *Player) SetFreeCameraEased(position, target mgl64.Vec3, seconds float32) {
	if p.session() != session.Nop {
		p.session().SetFreeCameraEased(position, target, seconds)
	}
}
func (p *Player) FadeCamera(in, hold, out float32) {
	if p.session() != session.Nop {
		p.session().FadeCamera(in, hold, out)
	}
}
func (p *Player) ClearCamera() {
	if p.session() != session.Nop {
		p.session().ClearCamera()
	}
}

func (p *Player) ShakeCamera(intensity, duration float32) {
	if p.session() != session.Nop {
		p.session().ShakeCamera(intensity, duration)
	}
}

// CanSeeEntity reports explicit hiding, independently of chunk visibility.
func (p *Player) CanSeeEntity(e world.Entity) bool {
	return p.session() == session.Nop || !p.session().EntityHidden(e)
}

// ViewCameraActor shows a camera subject independently of the player's chunk
// radius. Visibility restrictions on the actor still apply in the session.
func (p *Player) ViewCameraActor(e world.Entity) {
	if p.session() != session.Nop {
		p.session().ViewEntity(e)
		p.session().ViewEntityArmour(e)
	}
}
func (p *Player) ViewCameraArmour(e world.Entity) {
	if p.session() != session.Nop {
		p.session().ViewEntityArmour(e)
	}
}
func (p *Player) HideCameraActor(e world.Entity) {
	if p.session() != session.Nop {
		p.session().HideEntity(e)
	}
}
func (p *Player) SendActionBar(text string) {
	if p.session() != session.Nop {
		p.session().SendActionBarMessage(text)
	}
}
