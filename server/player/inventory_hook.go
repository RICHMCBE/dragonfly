package player

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
)

// AllowInventoryDestination permits plugins to reject storage before any stack
// movement occurs. UI slot zero is the cursor; crafting slots are not personal storage.
func (p *Player) AllowInventoryDestination(inv *inventory.Inventory, slot int, stack item.Stack) bool {
	personal := inv == p.inv || inv == p.offHand || inv == p.armour.Inventory() || (inv == p.ui && slot == 0)
	if hook, ok := p.Handler().(interface {
		AllowInventoryDestination(*Player, *inventory.Inventory, int, item.Stack, bool) bool
	}); ok {
		return hook.AllowInventoryDestination(p, inv, slot, stack, personal)
	}
	return true
}
func (p *Player) InventoryTransactionCommitted() {
	if hook, ok := p.Handler().(interface{ HandleInventoryCommit(*Player) }); ok {
		hook.HandleInventoryCommit(p)
	}
}
func (p *Player) keepDeathItem(stack item.Stack) bool {
	if hook, ok := p.Handler().(interface {
		KeepDeathItem(*Player, item.Stack) bool
	}); ok {
		return hook.KeepDeathItem(p, stack)
	}
	return false
}
