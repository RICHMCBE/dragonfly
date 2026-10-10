package item

// FishingRod leaves casting to the server's player handler so custom fishing
// sessions can retain their own bobber and reward rules.
type FishingRod struct{}

func (FishingRod) MaxCount() int               { return 1 }
func (FishingRod) EncodeItem() (string, int16) { return "minecraft:fishing_rod", 0 }
func (FishingRod) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{MaxDurability: 384, BrokenItem: simpleItem(Stack{}), AttackDurability: 0, BreakDurability: 0}
}
