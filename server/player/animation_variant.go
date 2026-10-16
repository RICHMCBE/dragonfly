package player

// Variant is a transient resource-pack animation selector. It is not persisted.
func (p *Player) Variant() int32 { return p.animationVariant }

// SetAnimationVariant publishes the selector to viewers, including the owning client.
func (p *Player) SetAnimationVariant(value int32) {
	if p.animationVariant == value {
		return
	}
	p.animationVariant = value
	p.updateState()
}
