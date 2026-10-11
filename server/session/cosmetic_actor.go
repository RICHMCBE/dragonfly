package session

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Cosmetic actors only exist in this viewer's connection. They are deliberately
// absent from entities, so interactions cannot damage, pick up or merge them.
func (s *Session) nextCosmeticActor() uint64 {
	s.entityMutex.Lock()
	defer s.entityMutex.Unlock()
	s.currentEntityRuntimeID++
	if s.cosmeticActors == nil {
		s.cosmeticActors = map[uint64]struct{}{}
	}
	s.cosmeticActors[s.currentEntityRuntimeID] = struct{}{}
	return s.currentEntityRuntimeID
}
func cosmeticMetadata(scale float32) protocol.EntityMetadata {
	m := protocol.NewEntityMetadata()
	m[protocol.EntityDataKeyScale] = scale
	m.SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagNoAI)
	return m
}
func (s *Session) ShowCosmeticActor(identifier string, pos mgl64.Vec3, yaw float64, scale float32, variant, mark int32) uint64 {
	id := s.nextCosmeticActor()
	metadata := cosmeticMetadata(scale)
	metadata[protocol.EntityDataKeyVariant] = variant
	metadata[protocol.EntityDataKeyMarkVariant] = mark
	s.writePacket(&packet.AddActor{EntityUniqueID: int64(id), EntityRuntimeID: id, EntityType: identifier, Position: vec64To32(pos), Yaw: float32(yaw), HeadYaw: float32(yaw), BodyYaw: float32(yaw), EntityMetadata: metadata})
	return id
}
func (s *Session) ShowCosmeticItem(stack item.Stack, pos mgl64.Vec3) uint64 {
	if stack.Empty() {
		return 0
	}
	id := s.nextCosmeticActor()
	s.writePacket(&packet.AddItemActor{EntityUniqueID: int64(id), EntityRuntimeID: id, Item: instanceFromItem(s.br, stack), Position: vec64To32(pos), EntityMetadata: cosmeticMetadata(1)})
	return id
}
func (s *Session) MoveCosmeticActor(id uint64, pos mgl64.Vec3, yaw float64) {
	s.entityMutex.RLock()
	_, ok := s.cosmeticActors[id]
	s.entityMutex.RUnlock()
	if !ok {
		return
	}
	s.writePacket(&packet.MoveActorAbsolute{EntityRuntimeID: id, Position: vec64To32(pos), Rotation: mgl32.Vec3{0, float32(yaw), 0}})
}
func (s *Session) HideCosmeticActor(id uint64) {
	s.entityMutex.Lock()
	_, ok := s.cosmeticActors[id]
	delete(s.cosmeticActors, id)
	s.entityMutex.Unlock()
	if !ok {
		return
	}
	s.writePacket(&packet.RemoveActor{EntityUniqueID: int64(id)})
}
