package player

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
)

// OpenInventory opens a native chest inventory, retaining native item movement.
// Callbacks run on the player's world owner, after request rollback/commit.
func (p *Player) OpenInventory(title string, inv *inventory.Inventory, close, afterRequest func(*Player)) error {
	if p.session() == session.Nop {
		return fmt.Errorf("player has no session")
	}
	pos := cube.PosFromVec3(p.Position())
	r := p.tx.Range()
	pos[1] = max(r[0], min(r[1]-1, pos[1]-3))
	h := p.H()
	wrap := func(f func(*Player)) func(*world.Tx) {
		if f == nil {
			return nil
		}
		return func(tx *world.Tx) {
			e, ok := h.Entity(tx)
			if ok {
				f(e.(*Player))
			}
		}
	}
	return p.session().OpenVirtualContainer(p.tx, pos, title, inv, wrap(close), wrap(afterRequest))
}
func (p *Player) CloseInventory() {
	if p.session() != session.Nop {
		p.session().CloseContainer(p.tx)
	}
}

func (p *Player) RefreshCreativeInventory() {
	if p.session() != session.Nop {
		p.session().SendCreativeContent()
	}
}

// TemporaryInventory contains cursor/crafting stacks which are not part of the
// persisted main inventory. Transactional container editors must checkpoint it.
func (p *Player) TemporaryInventory() *inventory.Inventory { return p.ui }
