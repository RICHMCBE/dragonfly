package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// redstoneInertBlock takes no part in redstone, like crops or farmland.
type redstoneInertBlock struct{ Stage int }

func (b redstoneInertBlock) EncodeBlock() (string, map[string]any) {
	return "test:inert", map[string]any{"stage": int32(b.Stage)}
}
func (b redstoneInertBlock) Hash() (uint64, uint64) { return 1 << 50, uint64(b.Stage) }
func (redstoneInertBlock) Model() BlockModel        { return redstoneCancellationModel{} }

// TestRedstoneSkipsChangesOutOfReach checks that a change with no redstone in the sub chunks within four
// blocks queues no redstone work, while every change within four blocks of redstone still does.
func TestRedstoneSkipsChangesOutOfReach(t *testing.T) {
	registry := redstoneCancellationTestRegistry()
	for stage := int32(0); stage < 2; stage++ {
		registry.RegisterBlockState(BlockState{Name: "test:inert", Properties: map[string]any{"stage": stage}})
		registry.RegisterBlock(redstoneInertBlock{Stage: int(stage)})
	}
	w := Config{Synchronous: true, Blocks: registry}.New()
	defer w.Close()
	consumer := cube.Pos{0, 64, 0}
	dirty := func(tx *Tx, pos cube.Pos, stage int) int {
		clear(tx.World().redstone.dirty)
		tx.SetBlock(pos, redstoneInertBlock{Stage: stage}, nil)
		return len(tx.World().redstone.dirty)
	}
	runWorld(w, func(tx *Tx) {
		if n := dirty(tx, cube.Pos{2, 64, 0}, 0); n != 0 {
			t.Fatalf("change with no redstone anywhere queued %d dirty positions, want 0", n)
		}
	})
	runWorld(w, func(tx *Tx) {
		tx.SetBlock(consumer, redstoneCancellationConsumer{}, &SetOpts{DisableRedstoneUpdates: true})
		for _, c := range []struct {
			pos  cube.Pos
			want bool
		}{
			{cube.Pos{21, 64, 0}, false}, // next chunk: its reach stops at x 17
			{cube.Pos{8, 90, 8}, false},  // sub chunk above: its reach stops at y 86
			{cube.Pos{-1, 62, 1}, true},  // reach spans the chunk and sub chunk boundaries
			{cube.Pos{4, 64, 0}, true},
			{cube.Pos{2, 65, 1}, true},
			{cube.Pos{1, 64, 0}, true},
		} {
			for stage := 0; stage < 2; stage++ {
				if got := dirty(tx, c.pos, stage) > 0; got != c.want {
					t.Errorf("change at %v (stage %d), consumer at %v: redstone queued %t, want %t", c.pos, stage, consumer, got, c.want)
				}
			}
		}
	})
}
