package skin

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type protocolSnapshot struct {
	data        []byte
	fingerprint [32]byte
}

func (s Skin) fingerprint() [32]byte {
	var animationShapes [][3]int
	for _, a := range s.Animations {
		animationShapes = append(animationShapes, [3]int{a.w, a.h, int(a.aType)})
	}
	data, _ := json.Marshal(struct {
		Skin                                 Skin
		Width, Height, CapeWidth, CapeHeight int
		AnimationShapes                      [][3]int
	}{s, s.w, s.h, s.Cape.Bounds().Dx(), s.Cape.Bounds().Dy(), animationShapes})
	return sha256.Sum256(data)
}

// PreserveProtocol retains an independent copy, never a cache keyed by SkinID.
// Server edits invalidate it, preventing stale profile hashes on custom skins.
func (s *Skin) PreserveProtocol(p protocol.Skin) {
	data, _ := json.Marshal(p)
	s.original = &protocolSnapshot{data: data, fingerprint: s.fingerprint()}
}

func (s Skin) OriginalProtocol() (protocol.Skin, bool) {
	if s.original == nil || s.fingerprint() != s.original.fingerprint {
		return protocol.Skin{}, false
	}
	var p protocol.Skin
	err := json.Unmarshal(s.original.data, &p)
	return p, err == nil
}
