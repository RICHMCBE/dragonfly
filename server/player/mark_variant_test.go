package player

import "testing"

// A Player value lasts one transaction, so the mark has to live in the data they share.
func TestMarkVariantOutlivesTheTransaction(t *testing.T) {
	data := &playerData{}
	(&Player{playerData: data}).markVariant = 101
	if got := (&Player{playerData: data}).MarkVariant(); got != 101 {
		t.Fatal(got)
	}
}
