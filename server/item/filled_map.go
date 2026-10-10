package item

// FilledMap preserves map identity and source art metadata through inventory,
// item frames and player storage. Pixels are supplied on MapInfoRequest.
type FilledMap struct {
	ID                          int64
	NameIndex, Scale            int32
	DisplayPlayers, Initialised bool
	Order                       string
}

func (FilledMap) EncodeItem() (string, int16) { return "minecraft:filled_map", 0 }
func (FilledMap) MaxCount() int               { return 64 }
func (FilledMap) OffHand() bool               { return true }
func (m FilledMap) EncodeNBT() map[string]any {
	var display, init byte
	if m.DisplayPlayers {
		display = 1
	}
	if m.Initialised {
		init = 1
	}
	v := map[string]any{"map_uuid": m.ID, "map_name_index": m.NameIndex, "map_display_players": display, "map_is_init": init, "map_scale": m.Scale}
	if m.Order != "" {
		v["poporand_art_order"] = m.Order
	}
	return v
}
func (m FilledMap) DecodeNBT(v map[string]any) any {
	m.ID, _ = v["map_uuid"].(int64)
	m.NameIndex, _ = v["map_name_index"].(int32)
	m.Scale, _ = v["map_scale"].(int32)
	display, _ := v["map_display_players"].(byte)
	init, _ := v["map_is_init"].(byte)
	m.DisplayPlayers = display != 0
	m.Initialised = init != 0
	m.Order, _ = v["poporand_art_order"].(string)
	return m
}
