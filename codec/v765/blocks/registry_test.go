package blocks_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec/v765/blocks"
)

func TestOfResolvesState(t *testing.T) {
	block, ok := blocks.Of(2885)
	if !ok {
		t.Fatal("no block for state 2885")
	}
	if block.Name != "oak_stairs" {
		t.Errorf("name = %q, want oak_stairs", block.Name)
	}
}

func TestNamedFindsRange(t *testing.T) {
	block, ok := blocks.Named("oak_stairs")
	if !ok {
		t.Fatal("oak_stairs not found")
	}
	if block.MinState != 2874 || block.MaxState != 2953 {
		t.Errorf("range = [%d, %d], want [2874, 2953]", block.MinState, block.MaxState)
	}
}

func TestNamedDecomposesState(t *testing.T) {
	block, ok := blocks.Named("oak_stairs")
	if !ok {
		t.Fatal("oak_stairs not found")
	}

	want := map[string]string{"facing": "north", "half": "top", "shape": "straight", "waterlogged": "true"}
	for name, value := range block.At(block.MinState) {
		if want[name] != value {
			t.Errorf("%s = %q, want %q", name, value, want[name])
		}
	}
}
