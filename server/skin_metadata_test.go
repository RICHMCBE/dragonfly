package server

import (
	"encoding/base64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"testing"
)

func TestLoginSkinMetadata(t *testing.T) {
	b64 := base64.StdEncoding.EncodeToString
	d := login.ClientData{SkinID: "login", ProfileHash: "profile", SkinImageWidth: 64, SkinImageHeight: 64, SkinData: b64(make([]byte, 64*64*4)), SkinGeometry: b64([]byte(`{}`)), SkinResourcePatch: b64([]byte(`{"geometry":{"default":"geometry.humanoid.custom"}}`)), SkinGeometryVersion: b64([]byte("1.26.50")), ArmSize: "slim", SkinColour: "#b37b62", CapeID: "cape", TrustedSkin: true}
	s := new(Server).parseSkin(d)
	p, ok := s.OriginalProtocol()
	if !ok || p.ProfileHash != d.ProfileHash || p.SkinID != d.SkinID || p.ArmSize != protocol.ArmSizeSlim || p.SkinColour.A != 255 || p.SkinColour.R != 0xb3 || string(p.GeometryDataEngineVersion) != "1.26.50" {
		t.Fatal("login metadata lost")
	}
}
