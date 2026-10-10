package session

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SetFreeCamera installs a server controlled free camera and points it at target.
func (s *Session) SetFreeCamera(position, target mgl64.Vec3) {
	s.writePacket(&packet.CameraPresets{Presets: []protocol.CameraPreset{{Name: "popoland:preview", Parent: "minecraft:free"}}})
	s.writePacket(&packet.CameraInstruction{Set: protocol.Option(protocol.CameraInstructionSet{Preset: 0, Position: protocol.Option(mgl32.Vec3{float32(position[0]), float32(position[1]), float32(position[2])}), Facing: protocol.Option(mgl32.Vec3{float32(target[0]), float32(target[1]), float32(target[2])})})})
}
func (s *Session) ClearCamera() {
	s.writePacket(&packet.CameraInstruction{Clear: protocol.Option(true)})
}
