package world

import (
	"context"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

type reentrantChunkViewer struct {
	NopViewer
	onChunk func(ChunkPos)
}

func (v reentrantChunkViewer) ViewChunk(pos ChunkPos, _ Dimension, _ map[cube.Pos]Block, _ *chunk.Chunk) {
	v.onChunk(pos)
}

func TestLoaderViewChunkAllowsNestedChunkCompletion(t *testing.T) {
	w := Config{}.New()
	first, second := &Column{}, &Column{}
	l := &Loader{r: 2, w: w, loaded: map[ChunkPos]*Column{}, pending: map[ChunkPos]struct{}{}}
	var tx *Tx
	count := 0
	l.viewer = reentrantChunkViewer{onChunk: func(pos ChunkPos) {
		count++
		if _, ok := l.Chunk(pos); !ok {
			panic("chunk must be published before viewer callbacks")
		}
		if pos == (ChunkPos{}) {
			l.viewChunk(tx, ChunkPos{1, 0}, second)
		}
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := w.Do(func(owner *Tx) {
		tx = owner
		l.viewChunk(owner, ChunkPos{}, first)
	}).Wait(ctx); err != nil {
		t.Fatalf("nested chunk completion stalled: %v", err)
	}
	defer w.Close()
	if count != 2 || len(l.loaded) != 2 || len(first.viewers) != 1 || len(second.viewers) != 1 {
		t.Fatalf("nested completion lost state: callbacks=%d loaded=%d", count, len(l.loaded))
	}
}
