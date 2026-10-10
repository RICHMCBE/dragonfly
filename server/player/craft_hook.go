package player

import "github.com/df-mc/dragonfly/server/item"

// AllowCraft invokes an optional pre-consumption crafting guard.
func (p *Player) AllowCraft(outputs []item.Stack) bool {
	if h, ok := p.Handler().(interface{ HandleCraft(*Context, []item.Stack) }); ok {
		ctx := NewEventContext(p.tx, p)
		h.HandleCraft(ctx, outputs)
		return !ctx.Cancelled()
	}
	return true
}
