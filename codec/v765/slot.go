package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

type Slot struct {
	Present codec.Bool
	Item    codec.VarInt
	Count   codec.Byte
	Data    codec.NBT
}

func slotOf(stack graft.ItemStack) Slot {
	if stack.Empty() {
		return Slot{}
	}

	return Slot{
		Present: true,
		Item:    codec.VarInt(stack.Item),
		Count:   codec.Byte(stack.Count),
		Data:    stack.Data,
	}
}

func (s Slot) Append(dst []byte) []byte {
	if !s.Present.Bool() {
		return s.Present.Append(dst)
	}

	return codec.AppendAll(dst, s.Present, s.Item, s.Count, s.Data)
}

func (s *Slot) Decode(r *codec.Reader) error {
	if err := s.Present.Decode(r); err != nil {
		return err
	}
	if !s.Present.Bool() {
		*s = Slot{}

		return nil
	}

	return codec.DecodeAll(r, &s.Item, &s.Count, &s.Data)
}

func (s Slot) Stack() graft.ItemStack {
	if !s.Present.Bool() {
		return graft.ItemStack{}
	}

	return graft.ItemStack{
		Item:  graft.ItemID(s.Item),
		Count: s.Count.Int(),
		Data:  s.Data,
	}
}

type ChangedSlot struct {
	Index codec.Short
	Item  Slot
}

func (c ChangedSlot) Append(dst []byte) []byte {
	return codec.AppendAll(dst, c.Index, c.Item)
}

func (c *ChangedSlot) Decode(r *codec.Reader) error {
	return codec.DecodeAll(r, &c.Index, &c.Item)
}
