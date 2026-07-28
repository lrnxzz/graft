package gocraft

type ItemID int32

const MaxStackSize = 64

type Item struct {
	ID        ItemID     `json:"id"`
	Name      Identifier `json:"name"`
	StackSize int        `json:"stackSize"`
}

type ItemStack struct {
	Item  ItemID
	Count int
	Data  NBT
}

func (s ItemStack) Empty() bool {
	return s.Item == 0 || s.Count <= 0
}

func (s ItemStack) Is(item ItemID) bool {
	return !s.Empty() && s.Item == item
}
