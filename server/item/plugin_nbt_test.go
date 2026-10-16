package item_test

import (
	"github.com/df-mc/dragonfly/server/item"
	"testing"
)

func TestPluginTagUpdatesDoNotDamageWornTool(t *testing.T) {
	s := item.NewStack(item.Pickaxe{Tier: item.ToolTierNetherite}, 1).WithDurability(2)
	for i := 0; i < 5; i++ {
		s = s.WithNBTValue("toolcore_test", int32(i))
		if s.Empty() || s.Durability() != 2 {
			t.Fatal("tag update applied damage again", i, s.Durability())
		}
	}
	s = s.WithNBTValue("toolcore_test", nil)
	if s.Durability() != 2 {
		t.Fatal(s.Durability())
	}
}
