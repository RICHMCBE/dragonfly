package session

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"image/color"
)

// SetFreeCamera installs a server controlled free camera and points it at target.
func (s *Session) SetFreeCamera(position, target mgl64.Vec3) {
	s.SetFreeCameraEased(position, target, 0)
}

func (s *Session) SetFreeCameraEased(position, target mgl64.Vec3, seconds float32) {
	// Register once: replacing presets during an active camera invalidates the
	// client's current preset. Include the native free preset explicitly.
	if !s.cameraPresetsSent {
		s.writePacket(&packet.CameraPresets{Presets: []protocol.CameraPreset{
			{Name: "minecraft:free"}, {Name: "minecraft:first_person"},
			{Name: "minecraft:third_person"}, {Name: "minecraft:third_person_front"},
		}})
		s.cameraPresetsSent = true
	}
	instruction := protocol.CameraInstructionSet{Preset: 0, Position: protocol.Option(mgl32.Vec3{float32(position[0]), float32(position[1]), float32(position[2])}), Facing: protocol.Option(mgl32.Vec3{float32(target[0]), float32(target[1]), float32(target[2])})}
	if seconds > 0 {
		instruction.Ease = protocol.Option(protocol.CameraEase{Type: protocol.EasingTypeInOutSine, Duration: seconds})
	}
	s.writePacket(&packet.CameraInstruction{Set: protocol.Option(instruction)})
}

func (s *Session) FadeCamera(in, hold, out float32) {
	s.writePacket(&packet.CameraInstruction{Fade: protocol.Option(protocol.CameraInstructionFade{
		TimeData: protocol.Option(protocol.CameraFadeTimeData{FadeInDuration: in, WaitDuration: hold, FadeOutDuration: out}),
		Colour:   protocol.Option(color.RGBA{R: 3, G: 9, B: 6, A: 255}),
	})})
}
func (s *Session) ClearCamera() {
	s.writePacket(&packet.CameraShake{Type: packet.CameraShakeTypeRotational, Action: packet.CameraShakeActionStop})
	s.writePacket(&packet.CameraInstruction{Clear: protocol.Option(true)})
}

// ShakeCamera adds a short rotational impulse without moving the player.
func (s *Session) ShakeCamera(intensity, duration float32) {
	s.writePacket(&packet.CameraShake{Intensity: intensity, Duration: duration, Type: packet.CameraShakeTypeRotational, Action: packet.CameraShakeActionAdd})
}
