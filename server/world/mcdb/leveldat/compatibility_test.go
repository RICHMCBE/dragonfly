package leveldat

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"testing"
)

func TestPMMPAdditionalFieldsSurviveSave(t *testing.T) {
	fields := map[string]any{"LevelName": "world", "randomTickSpeed": int32(3), "pmmpCustom": map[string]any{"value": int64(42)}}
	raw, err := nbt.MarshalEncoding(fields, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	ld := LevelDat{data: raw}
	var data Data
	if err = ld.Unmarshal(&data); err != nil {
		t.Fatal(err)
	}
	data.LevelName = "renamed"
	var saved LevelDat
	if err = saved.Marshal(data); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err = saved.Unmarshal(&result); err != nil {
		t.Fatal(err)
	}
	if result["LevelName"] != "renamed" || result["randomTickSpeed"] != int32(3) || result["pmmpCustom"].(map[string]any)["value"] != int64(42) {
		t.Fatal(result)
	}
}
