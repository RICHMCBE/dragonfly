package player

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// linkViewer is a world.Viewer that can show one entity riding another. The network session
// implements it; other viewers simply do not see the link.
type linkViewer interface {
	ViewEntityLink(rider, ridden world.Entity, linked bool)
}

// Mount seats the player on e, with the seat seat blocks away from e's position. The client
// draws the player sitting there until Dismount is called, the player sneaks, leaves the vehicle,
// teleports or dies. Mounting another entity first leaves the current one.
func (p *Player) Mount(e world.Entity, seat mgl64.Vec3) {
	if e == nil || e.H() == p.handle {
		return
	}
	if p.mount != nil {
		p.Dismount()
	}
	p.mount, p.seat = e.H(), seat
	p.teleport(e.Position().Add(seat))
	p.updateState()
	for _, v := range p.viewers() {
		if lv, ok := v.(linkViewer); ok {
			lv.ViewEntityLink(p, e, true)
		}
	}
}

// Dismount makes the player leave the entity it sits on, if any, and stand just above the seat.
func (p *Player) Dismount() {
	h := p.mount
	if h == nil {
		return
	}
	p.mount, p.seat = nil, mgl64.Vec3{}
	ridden, ok := h.Entity(p.tx)
	for _, v := range p.viewers() {
		if lv, ok2 := v.(linkViewer); ok2 && ok {
			lv.ViewEntityLink(p, ridden, false)
		}
	}
	p.updateState()
	if ok {
		p.teleport(ridden.Position().Add(mgl64.Vec3{0, 1, 0}))
	}
}

// Mounted returns the entity the player sits on.
func (p *Player) Mounted() (*world.EntityHandle, bool) {
	return p.mount, p.mount != nil
}

// SeatOffset returns where the player sits relative to the entity it rides, and false when it
// rides nothing. The network session puts it in the player's metadata.
func (p *Player) SeatOffset() (mgl64.Vec3, bool) {
	return p.seat, p.mount != nil
}
