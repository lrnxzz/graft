package blocks_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec/v765/blocks"
)

func TestGeneratedConstantsMatchRegistry(t *testing.T) {
	oak, ok := blocks.Named("oak_stairs")
	if !ok {
		t.Fatal("oak_stairs not found")
	}
	if blocks.OakStairs != oak.DefaultState {
		t.Errorf("OakStairs = %d, want %d", blocks.OakStairs, oak.DefaultState)
	}

	stone, ok := blocks.Named("stone")
	if !ok {
		t.Fatal("stone not found")
	}
	if blocks.Stone != stone.DefaultState {
		t.Errorf("Stone = %d, want %d", blocks.Stone, stone.DefaultState)
	}
}
