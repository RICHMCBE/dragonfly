package server

import (
	"encoding/base64"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"image/color"
	"strconv"
	"strings"
)

func preserveLoginSkin(s *skin.Skin, d login.ClientData) {
	decode := func(v string) []byte { b, _ := base64.StdEncoding.DecodeString(v); return b }
	p := protocol.Skin{
		SkinID: d.SkinID, FullID: d.SkinID, PlayFabID: d.PlayFabID,
		SkinResourcePatch: decode(d.SkinResourcePatch), SkinGeometry: s.Model,
		SkinImageWidth: uint32(d.SkinImageWidth), SkinImageHeight: uint32(d.SkinImageHeight), SkinData: s.Pix,
		CapeImageWidth: uint32(d.CapeImageWidth), CapeImageHeight: uint32(d.CapeImageHeight), CapeData: s.Cape.Pix, CapeID: d.CapeID,
		GeometryDataEngineVersion: decode(d.SkinGeometryVersion), AnimationData: decode(d.SkinAnimationData),
		PremiumSkin: d.PremiumSkin, PersonaSkin: d.PersonaSkin, PersonaCapeOnClassicSkin: d.CapeOnClassicSkin,
		Trusted: d.TrustedSkin, OverrideAppearance: d.OverrideSkin, ProfileHash: d.ProfileHash,
		SkinColour: loginSkinColour(d.SkinColour), ArmSize: protocol.ArmSizeWide,
	}
	if d.ArmSize == "slim" {
		p.ArmSize = protocol.ArmSizeSlim
	}
	for _, a := range d.AnimatedImageData {
		p.Animations = append(p.Animations, protocol.SkinAnimation{ImageWidth: uint32(a.ImageWidth), ImageHeight: uint32(a.ImageHeight), ImageData: decode(a.Image), AnimationType: uint32(a.Type), FrameCount: float32(a.Frames), ExpressionType: uint32(a.AnimationExpression)})
	}
	types := []string{"unknown", "skeleton", "body", "skin", "bottom", "feet", "dress", "top", "high_pants", "hands", "outerwear", "facial_hair", "mouth", "eyes", "hair", "hood", "back", "face_accessory", "head", "legs", "left_leg", "right_leg", "arms", "left_arm", "right_arm", "capes", "classic_skin", "emote", "unsupported"}
	for _, piece := range d.PersonaPieces {
		kind := strings.TrimPrefix(piece.PieceType, "persona_")
		if kind == "hand" {
			kind = "hands"
		}
		typ := uint32(protocol.PieceTypeUnknown)
		for i, name := range types {
			if name == kind {
				typ = uint32(i)
				break
			}
		}
		pack, _ := uuid.Parse(piece.PackID)
		p.PersonaPieces = append(p.PersonaPieces, protocol.PersonaPiece{PieceID: piece.PieceID, PieceType: typ, PackID: pack, Default: piece.Default, ProductID: piece.ProductID})
	}
	for _, t := range d.PieceTintColours {
		tint := protocol.PersonaPieceTintColour{PieceType: t.PieceType}
		for i, c := range t.Colours {
			tint.Colours[i] = loginSkinColour(c)
		}
		p.PieceTintColours = append(p.PieceTintColours, tint)
	}
	s.PreserveProtocol(p)
}

func loginSkinColour(s string) color.RGBA {
	raw := strings.TrimPrefix(s, "#")
	n, _ := strconv.ParseUint(raw, 16, 32)
	a := uint8(n >> 24)
	if len(raw) == 6 {
		a = 255
	}
	return color.RGBA{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n), A: a}
}
