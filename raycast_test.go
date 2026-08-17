package graft_test

import (
	"math"
	"testing"

	graft "github.com/lrnxzz/graft"
)

func TestRaycastHitsFacingBlock(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	column.SetBlock(0, 5, 3, 1)

	world := graft.NewWorld()
	world.LoadColumn(column)

	solid := func(state graft.BlockState) bool {
		return state != 0
	}

	hit, ok := world.Raycast(graft.Vec3(0.5, 5.5, 0.5), graft.Vec3(0, 0, 1), 10, solid)
	if !ok {
		t.Fatal("expected a hit")
	}

	want := graft.Position{
		X: 0,
		Y: 5,
		Z: 3,
	}
	if hit.Block != want {
		t.Errorf("hit block %v, want %v", hit.Block, want)
	}
	if hit.Face != graft.FaceNorth {
		t.Errorf("hit face %v, want north", hit.Face)
	}
	if math.Abs(hit.Distance-2.5) > 1e-9 {
		t.Errorf("hit distance %v, want 2.5", hit.Distance)
	}
	if math.Abs(hit.Point.Z-3) > 1e-9 {
		t.Errorf("hit point %v, want z=3", hit.Point)
	}
}

func TestRaycastHitsGroundFromAbove(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	column.SetBlock(4, 0, 4, 1)

	world := graft.NewWorld()
	world.LoadColumn(column)

	solid := func(state graft.BlockState) bool {
		return state != 0
	}

	hit, ok := world.Raycast(graft.Vec3(4.5, 3, 4.5), graft.Vec3(0, -1, 0), 10, solid)
	if !ok {
		t.Fatal("expected a hit")
	}

	if hit.Face != graft.FaceUp {
		t.Errorf("hit face %v, want up", hit.Face)
	}
	if math.Abs(hit.Distance-2) > 1e-9 {
		t.Errorf("hit distance %v, want 2", hit.Distance)
	}
}

func TestRaycastRespectsReach(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	column.SetBlock(0, 5, 3, 1)

	world := graft.NewWorld()
	world.LoadColumn(column)

	solid := func(state graft.BlockState) bool {
		return state != 0
	}

	_, ok := world.Raycast(graft.Vec3(0.5, 5.5, 0.5), graft.Vec3(0, 0, 1), 2, solid)
	if ok {
		t.Error("hit a block beyond reach")
	}
}

func TestRaycastStopsAtUnloadedChunks(t *testing.T) {
	world := graft.NewWorld()
	world.LoadColumn(graft.ChunkColumn(0, 0, -64, 384))

	solid := func(state graft.BlockState) bool {
		return state != 0
	}

	_, ok := world.Raycast(graft.Vec3(8, 5.5, 8), graft.Vec3(1, 0, 0), 64, solid)
	if ok {
		t.Error("hit a block inside an unloaded chunk")
	}
}
