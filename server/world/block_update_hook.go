package world

import "github.com/df-mc/dragonfly/server/block/cube"

// BlockUpdateHandler is an optional compatibility hook for cancellable block
// updates. Existing handlers keep their original behaviour when absent.
type BlockUpdateHandler interface {
	HandleBlockUpdate(ctx *Context, pos cube.Pos)
}

func allowBlockUpdate(tx *Tx, pos cube.Pos) bool {
	if h, ok := tx.World().Handler().(BlockUpdateHandler); ok {
		ctx := tx.Event()
		h.HandleBlockUpdate(ctx, pos)
		return !ctx.Cancelled()
	}
	return true
}
