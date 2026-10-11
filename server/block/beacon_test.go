package block_test

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestBeaconEffectNBT(t *testing.T) {
	for _, tc := range []struct {
		name               string
		data               map[string]any
		primary, secondary effect.LastingType
	}{
		{"bedrock", map[string]any{"primary": int32(1), "secondary": int32(10)}, effect.Speed, effect.Regeneration},
		{"legacy_dragonfly", map[string]any{"Primary": int32(3), "Secondary": int32(3)}, effect.Haste, effect.Haste},
		{"bedrock_precedence", map[string]any{"primary": int32(0), "Primary": int32(1), "secondary": int32(0), "Secondary": int32(10)}, nil, nil},
		{"invalid", map[string]any{"primary": int32(6), "secondary": int32(999)}, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := (block.Beacon{}).DecodeNBT(tc.data).(block.Beacon)
			if b.Primary != tc.primary || b.Secondary != tc.secondary {
				t.Fatalf("effects = %v, %v; want %v, %v", b.Primary, b.Secondary, tc.primary, tc.secondary)
			}
			data := b.EncodeNBT()
			for key, typ := range map[string]effect.LastingType{"primary": tc.primary, "secondary": tc.secondary} {
				id, _ := effect.ID(typ)
				if data[key] != int32(id) {
					t.Errorf("%s = %v; want %d", key, data[key], id)
				}
			}
		})
	}
}

func TestBedrockBeaconAppliesEffects(t *testing.T) {
	w := (world.Config{Synchronous: true, Entities: entity.DefaultRegistry}).New()
	defer w.Close()
	if err := w.Do(func(tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		for x := -1; x <= 1; x++ {
			for z := -1; z <= 1; z++ {
				tx.SetBlock(cube.Pos{x, 63, z}, block.Iron{}, nil)
			}
		}
		b := (block.Beacon{}).DecodeNBT(map[string]any{"primary": int32(1)}).(block.Beacon)
		tx.SetBlock(pos, b, nil)
		h := (world.EntitySpawnOpts{}).New(player.Type, player.Config{Position: mgl64.Vec3{2, 64, 2}})
		p := tx.AddEntity(h).(*player.Player)
		b.Tick(80, pos, tx)
		if got := tx.Block(pos).(block.Beacon).Level(); got != 1 {
			t.Fatalf("beacon level = %d; want 1", got)
		}
		if eff, ok := p.Effect(effect.Speed); !ok || eff.Level() != 1 {
			t.Fatalf("speed I not applied: effect=%v, present=%v", eff, ok)
		}
		for _, tc := range []struct {
			name   string
			roof   world.Block
			active bool
		}{
			{"opaque_roof", block.Stone{}, false},
			{"glass_roof", block.Glass{}, true},
			{"roof_removed", block.Air{}, true},
		} {
			p.RemoveEffect(effect.Speed)
			tx.SetBlock(pos.Side(cube.FaceUp), tc.roof, nil)
			tx.Block(pos).(block.Beacon).Tick(160, pos, tx)
			if _, ok := p.Effect(effect.Speed); ok != tc.active {
				t.Errorf("%s: speed present = %v; want %v", tc.name, ok, tc.active)
			}
		}
		tx.RemoveEntity(p)
		if err := h.Close(); err != nil {
			t.Fatal(err)
		}
	}).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
