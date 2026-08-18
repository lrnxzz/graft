package mesh

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/viewer/gpu"
)

// each state maps to whether it fills its cube opaquely
type fakeBlocks map[graft.BlockState]bool

func (f fakeBlocks) Tile(graft.BlockState, Face) gpu.UV { return gpu.UV{} }
func (f fakeBlocks) Solid(state graft.BlockState) bool  { return f[state] }

// a quad is four vertices of six floats each
func quads(g Geometry) int {
	return len(g.vertices) / 24
}

func meshed(blocks Blocks, place func(*graft.World)) Geometry {
	world := graft.NewWorld()
	column := graft.ChunkColumn(0, 0, 0, 16)
	world.LoadColumn(column)
	place(world)

	return Chunk(world, column, blocks)
}

const (
	stone = graft.BlockState(1)
	grass = graft.BlockState(2)
	water = graft.BlockState(3)
)

var kinds = fakeBlocks{stone: true, grass: false, water: false}

func TestALoneBlockShowsEverySide(t *testing.T) {
	geometry := meshed(kinds, func(w *graft.World) {
		w.SetBlock(0, 0, 0, stone)
	})

	if got := quads(geometry); got != 6 {
		t.Errorf("a lone block meshed %d faces, want 6", got)
	}
}

// the bug this guards: grass over the ground used to cull the ground's top face,
// leaving a hole that showed the sky
func TestGrassDoesNotHideTheGroundBehindIt(t *testing.T) {
	geometry := meshed(kinds, func(w *graft.World) {
		w.SetBlock(0, 0, 0, stone)
		w.SetBlock(0, 1, 0, grass)
	})

	// the stone keeps all six faces, the grass loses only the side against stone
	if got := quads(geometry); got != 11 {
		t.Errorf("stone under grass meshed %d faces, want 11 (6 stone + 5 grass)", got)
	}
}

func TestOpaqueNeighboursHideTheSharedFace(t *testing.T) {
	geometry := meshed(kinds, func(w *graft.World) {
		w.SetBlock(0, 0, 0, stone)
		w.SetBlock(0, 1, 0, stone)
	})

	if got := quads(geometry); got != 10 {
		t.Errorf("two stacked stones meshed %d faces, want 10", got)
	}
}

// water against water culls, or an ocean would mesh every interior face
func TestSameBlockCullsToo(t *testing.T) {
	geometry := meshed(kinds, func(w *graft.World) {
		w.SetBlock(0, 0, 0, water)
		w.SetBlock(0, 1, 0, water)
	})

	if got := quads(geometry); got != 10 {
		t.Errorf("two stacked water meshed %d faces, want 10", got)
	}
}
