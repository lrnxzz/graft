package items_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec/v765/items"
)

func TestOfResolvesID(t *testing.T) {
	item, ok := items.Of(items.Stone)
	if !ok {
		t.Fatal("no item for stone")
	}
	if item.Name != "stone" {
		t.Errorf("name = %q, want stone", item.Name)
	}
	if item.StackSize != 64 {
		t.Errorf("stack size = %d, want 64", item.StackSize)
	}
}

func TestNamedResolves(t *testing.T) {
	item, ok := items.Named("diamond_pickaxe")
	if !ok {
		t.Fatal("diamond_pickaxe not found")
	}
	if item.ID != items.DiamondPickaxe {
		t.Errorf("id = %d, want %d", item.ID, items.DiamondPickaxe)
	}
	if item.StackSize != 1 {
		t.Errorf("stack size = %d, want 1", item.StackSize)
	}
}
