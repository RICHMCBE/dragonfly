package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/player/skin"
)

func TestSkinToProtocolDropsAnimationsWithMismatchedPixels(t *testing.T) {
	s := skin.New(64, 64)
	valid := skin.NewAnimation(32, 32, 0, skin.AnimationBody32x32)
	valid.Pix = make([]byte, 32*32*4)
	valid.FrameCount = 1
	broken := skin.NewAnimation(0, 0, 0, skin.AnimationHead)
	broken.Pix = make([]byte, 8192)
	s.Animations = []skin.Animation{broken, valid}
	got := skinToProtocol(s)
	if len(got.Animations) != 1 || got.Animations[0].ImageWidth != 32 {
		t.Fatalf("animations = %+v", got.Animations)
	}
}
