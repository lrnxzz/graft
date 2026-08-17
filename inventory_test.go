package graft_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
)

func TestInventoryLoadAndSlotBounds(t *testing.T) {
	stacks := make([]graft.ItemStack, graft.InventorySize)
	stacks[graft.SlotMainStart] = graft.ItemStack{
		Item:  7,
		Count: 3,
	}

	var inventory graft.Inventory
	inventory.Load(stacks)

	got := inventory.Slot(graft.SlotMainStart)
	if !got.Is(7) || got.Count != 3 {
		t.Errorf("slot = %+v, want item 7 count 3", got)
	}
	if !inventory.Slot(-1).Empty() || !inventory.Slot(graft.InventorySize).Empty() {
		t.Error("out of range slots should read as empty")
	}
}

func TestInventoryFindPrefersHotbar(t *testing.T) {
	var inventory graft.Inventory
	inventory.SetSlot(graft.SlotMainStart, graft.ItemStack{
		Item:  5,
		Count: 1,
	})
	inventory.SetSlot(graft.HotbarSlot(4), graft.ItemStack{
		Item:  5,
		Count: 1,
	})

	slot, ok := inventory.FindItem(5)
	if !ok {
		t.Fatal("item should be found")
	}
	if slot != graft.HotbarSlot(4) {
		t.Errorf("found slot %d, want hotbar slot %d", slot, graft.HotbarSlot(4))
	}
}

func TestInventoryCountSumsStorage(t *testing.T) {
	var inventory graft.Inventory
	inventory.SetSlot(graft.SlotMainStart, graft.ItemStack{
		Item:  9,
		Count: 30,
	})
	inventory.SetSlot(graft.HotbarSlot(0), graft.ItemStack{
		Item:  9,
		Count: 12,
	})
	inventory.SetSlot(graft.SlotOffhand, graft.ItemStack{
		Item:  9,
		Count: 1,
	})
	inventory.SetSlot(graft.SlotHead, graft.ItemStack{
		Item:  9,
		Count: 1,
	})

	if got := inventory.Count(9); got != 43 {
		t.Errorf("count = %d, want 43 (armor slots are not storage)", got)
	}
}

func TestInventorySwapAndHeld(t *testing.T) {
	var inventory graft.Inventory
	inventory.SetSlot(graft.SlotMainStart, graft.ItemStack{
		Item:  2,
		Count: 1,
	})

	inventory.Swap(graft.SlotMainStart, graft.HotbarSlot(0))
	if !inventory.Hotbar(0).Is(2) || !inventory.Slot(graft.SlotMainStart).Empty() {
		t.Error("swap should move the stack into the hotbar")
	}

	inventory.SelectHeld(0)
	if !inventory.Held().Is(2) {
		t.Error("held stack should follow the selected hotbar index")
	}

	inventory.SelectHeld(9)
	if inventory.HeldIndex() != 0 {
		t.Error("out of range held index should be ignored")
	}
}

func TestInventoryFirstEmpty(t *testing.T) {
	var inventory graft.Inventory
	for index := graft.SlotHotbarStart; index <= graft.SlotOffhand; index++ {
		inventory.SetSlot(index, graft.ItemStack{
			Item:  1,
			Count: 1,
		})
	}

	slot, ok := inventory.FirstEmpty()
	if !ok {
		t.Fatal("main storage should still have room")
	}
	if slot != graft.SlotMainStart {
		t.Errorf("first empty = %d, want %d", slot, graft.SlotMainStart)
	}
}
