package player

import "image/color"

func (p *Player) RequestMap(id int64) {
	if h, ok := p.Handler().(interface{ HandleMapRequest(*Player, int64) }); ok {
		h.HandleMapRequest(p, id)
	}
}
func (p *Player) SendMap(id int64, pixels []color.RGBA) { p.session().SendMap(id, pixels) }
