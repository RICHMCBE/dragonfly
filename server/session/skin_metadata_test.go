package session

import (
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"reflect"
	"testing"
)

func TestSkinProtocolMetadataRoundTrip(t *testing.T) {
	for _, persona := range []bool{false, true} {
		p := protocol.Skin{SkinID: "shared-id", FullID: "full", PlayFabID: "fab", ProfileHash: "profile", SkinImageWidth: 64, SkinImageHeight: 64, SkinData: make([]byte, 64*64*4), SkinGeometry: []byte(`{}`), SkinResourcePatch: []byte(`{"geometry":{"default":"geometry.humanoid.custom"}}`), ArmSize: protocol.ArmSizeSlim, PersonaSkin: persona, GeometryDataEngineVersion: []byte("1.26.50"), AnimationData: []byte("animation"), PremiumSkin: true, CapeID: "cape-id", Trusted: true}
		s, err := protocolToSkin(p)
		if err != nil {
			t.Fatal(err)
		}
		if got := skinToProtocol(s); !reflect.DeepEqual(got, p) {
			t.Fatalf("metadata lost: %+v", got)
		}
		other := p
		other.ProfileHash = "other"
		s2, err := protocolToSkin(other)
		if err != nil {
			t.Fatal(err)
		}
		if skinToProtocol(s2).ProfileHash == skinToProtocol(s).ProfileHash {
			t.Fatal("SkinID collision")
		}
		s.Pix[0] = 123
		if skinToProtocol(s).ProfileHash != "" {
			t.Fatal("edited skin retained stale profile")
		}
		if skinToProtocol(skin.New(64, 64)).ProfileHash != "" {
			t.Fatal("new skin inherited metadata")
		}
	}
}
