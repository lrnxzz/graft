package graft

import "sync"

type World struct {
	mu      sync.RWMutex
	columns map[ChunkPos]*Column
}

func NewWorld() *World {
	return &World{columns: make(map[ChunkPos]*Column)}
}

func (w *World) LoadColumn(c *Column) {
	w.mu.Lock()
	w.columns[c.Pos()] = c
	w.mu.Unlock()
}

func (w *World) UnloadColumn(at ChunkPos) {
	w.mu.Lock()
	delete(w.columns, at)
	w.mu.Unlock()
}

func (w *World) Column(at ChunkPos) (*Column, bool) {
	w.mu.RLock()
	c, ok := w.columns[at]
	w.mu.RUnlock()

	return c, ok
}

func (w *World) Loaded() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return len(w.columns)
}

// Surrounds reports whether every column within radius of the block has arrived,
// which is what a planner needs before it can read the terrain around a player
func (w *World) Surrounds(p Position, radius int) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	center := p.Chunk()
	span := int32(radius)

	for x := -span; x <= span; x++ {
		for z := -span; z <= span; z++ {
			_, loaded := w.columns[center.Offset(x, z)]
			if !loaded {
				return false
			}
		}
	}

	return true
}

func (w *World) Columns() []*Column {
	w.mu.RLock()
	defer w.mu.RUnlock()

	columns := make([]*Column, 0, len(w.columns))
	for _, column := range w.columns {
		columns = append(columns, column)
	}

	return columns
}

func (w *World) BlockAt(p Position) (BlockState, bool) {
	return w.Block(p.X, p.Y, p.Z)
}

func (w *World) Block(x, y, z int) (BlockState, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	c, ok := w.columns[At(x, y, z).Chunk()]
	if !ok || !c.Covers(y) {
		return 0, false
	}

	return c.Block(x&15, y, z&15), true
}

func (w *World) SetBlock(x, y, z int, state BlockState) {
	w.mu.Lock()
	defer w.mu.Unlock()

	c, ok := w.columns[At(x, y, z).Chunk()]
	if !ok || !c.Covers(y) {
		return
	}

	c.SetBlock(x&15, y, z&15, state)
}
