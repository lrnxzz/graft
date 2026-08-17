package blocks_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/codec/v765/items"
)

func TestBreakTicksStoneByHand(t *testing.T) {
	ticks, ok := blocks.BreakTicks(blocks.Stone, items.Air)
	if !ok {
		t.Fatal("stone should be breakable")
	}
	if ticks != 150 {
		t.Errorf("ticks = %d, want 150 (1.5 hardness without the right tool)", ticks)
	}
}

func TestBreakTicksStoneWithWoodenPickaxe(t *testing.T) {
	ticks, ok := blocks.BreakTicks(blocks.Stone, items.WoodenPickaxe)
	if !ok {
		t.Fatal("stone should be breakable")
	}
	if ticks != 23 {
		t.Errorf("ticks = %d, want 23 (speed 2 harvesting 1.5 hardness)", ticks)
	}
}

func TestBreakTicksDirtByHand(t *testing.T) {
	ticks, ok := blocks.BreakTicks(blocks.Dirt, items.Air)
	if !ok {
		t.Fatal("dirt should be breakable")
	}
	if ticks != 15 {
		t.Errorf("ticks = %d, want 15 (0.5 hardness, no tool required)", ticks)
	}
}

func TestBreakTicksBedrockIsUnbreakable(t *testing.T) {
	if _, ok := blocks.BreakTicks(blocks.Bedrock, items.NetheritePickaxe); ok {
		t.Error("bedrock should not be breakable")
	}
}
