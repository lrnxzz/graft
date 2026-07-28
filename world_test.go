package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestWorldTracksLoadedColumns(t *testing.T) {
	world := gocraft.NewWorld()
	world.LoadColumn(gocraft.ChunkColumn(1, -1, -64, 384))

	at := gocraft.Chunk(1, -1)

	_, loaded := world.Column(at)
	if !loaded {
		t.Fatalf("column %v not loaded", at)
	}
	if world.Loaded() != 1 {
		t.Errorf("Loaded() = %d, want 1", world.Loaded())
	}

	world.UnloadColumn(at)
	_, stillThere := world.Column(at)
	if stillThere {
		t.Fatal("column still present after unload")
	}
}

func TestWorldResolvesBlocksAcrossChunkBounds(t *testing.T) {
	world := gocraft.NewWorld()
	world.LoadColumn(gocraft.ChunkColumn(1, -1, -64, 384))

	world.SetBlock(20, 70, -5, 42)

	got, ok := world.Block(20, 70, -5)
	if !ok || got != 42 {
		t.Errorf("Block(20, 70, -5) = %d, %v, want 42, true", got, ok)
	}

	if _, ok := world.Block(0, 70, 0); ok {
		t.Error("Block in unloaded column returned ok")
	}
}
