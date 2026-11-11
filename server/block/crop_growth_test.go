package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// grows skips reading the surroundings only for rolls that could never pass, so it must agree
// with comparing the roll against CalculateGrowthChance for every layout and roll.
func TestCropGrowsMatchesGrowthChance(t *testing.T) {
	pos := cube.Pos{0, 65, 0}
	layouts := []struct {
		name       string
		dry        bool
		neighbours []cube.Pos
		best       bool
	}{
		{name: "hydrated alone", best: true},
		{name: "dry alone", dry: true},
		{name: "row", neighbours: []cube.Pos{{1, 65, 0}, {-1, 65, 0}}},
		{name: "cross", neighbours: []cube.Pos{{1, 65, 0}, {0, 65, 1}}},
		{name: "diagonal", neighbours: []cube.Pos{{1, 65, 1}}},
	}
	for _, layout := range layouts {
		t.Run(layout.name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()
			runWorld(w, func(tx *world.Tx) {
				hydration := 7
				if layout.dry {
					hydration = 0
				}
				for x := -1; x <= 1; x++ {
					for z := -1; z <= 1; z++ {
						tx.SetBlock(cube.Pos{x, 64, z}, Farmland{Hydration: hydration}, nil)
					}
				}
				tx.SetBlock(pos, WheatSeeds{}, nil)
				for _, n := range layout.neighbours {
					tx.SetBlock(n, WheatSeeds{}, nil)
				}
				c := WheatSeeds{}.crop
				chance := c.CalculateGrowthChance(pos, tx)
				if chance > maxGrowthChance {
					t.Fatalf("chance %v above maxGrowthChance %v", chance, maxGrowthChance)
				}
				if layout.best && chance != maxGrowthChance {
					t.Fatalf("best layout chance %v, want maxGrowthChance %v", chance, maxGrowthChance)
				}
				for i := 0; i <= 1000; i++ {
					roll := float64(i) / 1000
					if got, want := c.grows(roll, pos, tx), roll <= chance; got != want {
						t.Fatalf("roll %v: grows %v, want %v (chance %v)", roll, got, want, chance)
					}
				}
			})
		})
	}
}

// Watered farmland that is already fully hydrated is left as it is instead of being set again.
func TestHydratedFarmlandRandomTickKeepsBlock(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	pos := cube.Pos{0, 64, 0}
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(pos, Farmland{Hydration: 7}, nil)
		tx.SetBlock(cube.Pos{2, 64, 0}, Water{Still: true, Depth: 8}, nil)
		Farmland{Hydration: 7}.RandomTick(pos, tx, nil)
		if f, ok := tx.Block(pos).(Farmland); !ok || f.Hydration != 7 {
			t.Fatalf("farmland = %#v, want hydration 7", tx.Block(pos))
		}
		tx.SetBlock(pos, Farmland{Hydration: 3}, nil)
		Farmland{Hydration: 3}.RandomTick(pos, tx, nil)
		if f, ok := tx.Block(pos).(Farmland); !ok || f.Hydration != 7 {
			t.Fatalf("watered farmland = %#v, want hydration 7", tx.Block(pos))
		}
	})
}
