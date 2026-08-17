package graft

const (
	SlotCraftingOutput = 0
	SlotHead           = 5
	SlotChest          = 6
	SlotLegs           = 7
	SlotFeet           = 8
	SlotMainStart      = 9
	SlotHotbarStart    = 36
	SlotOffhand        = 45

	InventorySize = 46
	HotbarSize    = 9
)

func HotbarSlot(index int) int {
	return SlotHotbarStart + index
}

type Inventory struct {
	slots [InventorySize]ItemStack
	held  int
}

func inRange(index int) bool {
	return index >= 0 && index < InventorySize
}

func (i *Inventory) Slot(index int) ItemStack {
	if !inRange(index) {
		return ItemStack{}
	}

	return i.slots[index]
}

func (i *Inventory) SetSlot(index int, stack ItemStack) {
	if !inRange(index) {
		return
	}

	i.slots[index] = stack
}

func (i *Inventory) Load(stacks []ItemStack) {
	for index := range min(len(stacks), InventorySize) {
		i.slots[index] = stacks[index]
	}
}

func (i *Inventory) Swap(a, b int) {
	if !inRange(a) || !inRange(b) {
		return
	}

	i.slots[a], i.slots[b] = i.slots[b], i.slots[a]
}

func (i *Inventory) HeldIndex() int {
	return i.held
}

func (i *Inventory) SelectHeld(index int) {
	if index < 0 || index >= HotbarSize {
		return
	}

	i.held = index
}

func (i *Inventory) Held() ItemStack {
	return i.slots[HotbarSlot(i.held)]
}

func (i *Inventory) Hotbar(index int) ItemStack {
	if index < 0 || index >= HotbarSize {
		return ItemStack{}
	}

	return i.slots[HotbarSlot(index)]
}

func (i *Inventory) Offhand() ItemStack {
	return i.slots[SlotOffhand]
}

func (i *Inventory) Find(match func(ItemStack) bool) (int, bool) {
	for index := SlotHotbarStart; index <= SlotOffhand; index++ {
		if match(i.slots[index]) {
			return index, true
		}
	}
	for index := SlotMainStart; index < SlotHotbarStart; index++ {
		if match(i.slots[index]) {
			return index, true
		}
	}

	return 0, false
}

func (i *Inventory) FindItem(item ItemID) (int, bool) {
	matches := func(stack ItemStack) bool {
		return stack.Is(item)
	}

	return i.Find(matches)
}

func (i *Inventory) FirstEmpty() (int, bool) {
	empty := func(stack ItemStack) bool {
		return stack.Empty()
	}

	return i.Find(empty)
}

func (i *Inventory) Count(item ItemID) int {
	total := 0
	for index := SlotMainStart; index <= SlotOffhand; index++ {
		if i.slots[index].Is(item) {
			total += i.slots[index].Count
		}
	}

	return total
}
