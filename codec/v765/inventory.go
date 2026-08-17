package v765

import (
	"fmt"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/codec/v765/items"
)

func (s *Session) Inventory() *graft.Inventory {
	return &s.inventory
}

func (s *Session) onContainerContent(c *codec.Client, p *SetContainerContent) error {
	if p.WindowID != 0 {
		return nil
	}

	s.stateID = p.StateID.Int32()

	stacks := make([]graft.ItemStack, len(p.Slots))
	for index, slot := range p.Slots {
		stacks[index] = slot.Stack()
	}
	s.inventory.Load(stacks)
	s.carried = p.Carried.Stack()

	return nil
}

func (s *Session) onContainerSlot(c *codec.Client, p *SetContainerSlot) error {
	s.stateID = p.StateID.Int32()

	switch {
	case p.WindowID == -1 && p.Index == -1:
		s.carried = p.Data.Stack()
	case p.WindowID == 0:
		s.inventory.SetSlot(p.Index.Int(), p.Data.Stack())
	}

	return nil
}

func (s *Session) onHeldItem(c *codec.Client, p *SetHeldItem) error {
	s.inventory.SelectHeld(p.Slot.Int())

	return nil
}

func (s *Session) SelectHotbar(index int) error {
	if index < 0 || index >= graft.HotbarSize {
		return fmt.Errorf("v765: hotbar index %d is out of range", index)
	}

	if err := s.client.Send(&HeldItemChange{Slot: codec.Short(index)}); err != nil {
		return err
	}

	s.inventory.SelectHeld(index)

	return nil
}

func (s *Session) SwapWithHotbar(slot, hotbar int) error {
	if hotbar < 0 || hotbar >= graft.HotbarSize {
		return fmt.Errorf("v765: hotbar index %d is out of range", hotbar)
	}

	return s.swap(slot, codec.Byte(hotbar), graft.HotbarSlot(hotbar))
}

func (s *Session) SwapWithOffhand(slot int) error {
	return s.swap(slot, offhandButton, graft.SlotOffhand)
}

func (s *Session) swap(slot int, button codec.Byte, other int) error {
	if slot < 0 || slot >= graft.InventorySize {
		return fmt.Errorf("v765: inventory slot %d is out of range", slot)
	}
	if slot == other {
		return nil
	}

	first := s.inventory.Slot(slot)
	second := s.inventory.Slot(other)

	click := &ClickContainer{
		StateID: codec.VarInt(s.stateID),
		Index:   codec.Short(slot),
		Button:  button,
		Mode:    clickSwap,
		Changed: codec.Slice[ChangedSlot]{
			{
				Index: codec.Short(slot),
				Item:  slotOf(second),
			},
			{
				Index: codec.Short(other),
				Item:  slotOf(first),
			},
		},
		Carried: slotOf(s.carried),
	}

	if err := s.client.Send(click); err != nil {
		return err
	}

	s.inventory.Swap(slot, other)

	return nil
}

func (s *Session) Carried() graft.ItemStack {
	return s.carried
}

func (s *Session) ClickSlot(slot int) error {
	if slot < 0 || slot >= graft.InventorySize {
		return fmt.Errorf("v765: inventory slot %d is out of range", slot)
	}
	if slot == graft.SlotCraftingOutput && !s.carried.Empty() {
		return nil
	}

	current := s.inventory.Slot(slot)
	if current.Empty() && s.carried.Empty() {
		return nil
	}

	landed, carried := land(current, s.carried)

	changed := ChangedSlot{
		Index: codec.Short(slot),
		Item:  slotOf(landed),
	}
	click := &ClickContainer{
		StateID: codec.VarInt(s.stateID),
		Index:   codec.Short(slot),
		Mode:    clickPickup,
		Changed: codec.Slice[ChangedSlot]{changed},
		Carried: slotOf(carried),
	}
	if err := s.client.Send(click); err != nil {
		return err
	}

	s.inventory.SetSlot(slot, landed)
	s.carried = carried

	return nil
}

func land(slot, carried graft.ItemStack) (graft.ItemStack, graft.ItemStack) {
	if carried.Empty() || slot.Empty() || slot.Item != carried.Item {
		return carried, slot
	}

	size := graft.MaxStackSize
	item, known := items.Of(slot.Item)
	if known {
		size = item.StackSize
	}

	total := slot.Count + carried.Count
	slot.Count = min(total, size)
	carried.Count = total - slot.Count
	if carried.Count == 0 {
		carried = graft.ItemStack{}
	}

	return slot, carried
}
