package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/nbt"
)

func TestSlotRoundTripsThroughContainerSlot(t *testing.T) {
	original := &v765.SetContainerSlot{
		WindowID: 0,
		StateID:  7,
		Index:    36,
		Data: v765.Slot{
			Present: true,
			Item:    276,
			Count:   1,
			Data: codec.NBT{
				"Damage": nbt.Int(3),
			},
		},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("set container slot not registered")
	}

	got := decoded.(*v765.SetContainerSlot)
	if !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}

	stack := got.Data.Stack()
	if !stack.Is(276) || stack.Count != 1 {
		t.Errorf("stack = %+v, want item 276 count 1", stack)
	}
}

func TestEmptySlotCarriesNoPayload(t *testing.T) {
	original := &v765.SetContainerSlot{
		WindowID: 0,
		StateID:  1,
		Index:    9,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("set container slot not registered")
	}

	got := decoded.(*v765.SetContainerSlot)
	if !got.Data.Stack().Empty() {
		t.Errorf("stack = %+v, want empty", got.Data.Stack())
	}
}

func TestSetContainerContentLoadsEverySlot(t *testing.T) {
	slots := make(codec.Slice[v765.Slot], gocraft.InventorySize)
	slots[gocraft.SlotHotbarStart] = v765.Slot{
		Present: true,
		Item:    1,
		Count:   64,
	}

	original := &v765.SetContainerContent{
		WindowID: 0,
		StateID:  3,
		Slots:    slots,
		Carried: v765.Slot{
			Present: true,
			Item:    2,
			Count:   16,
		},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("set container content not registered")
	}

	got := decoded.(*v765.SetContainerContent)
	if len(got.Slots) != gocraft.InventorySize {
		t.Fatalf("decoded %d slots, want %d", len(got.Slots), gocraft.InventorySize)
	}
	if !got.Slots[gocraft.SlotHotbarStart].Stack().Is(1) {
		t.Errorf("hotbar slot = %+v, want item 1", got.Slots[gocraft.SlotHotbarStart].Stack())
	}
	if !got.Carried.Stack().Is(2) {
		t.Errorf("carried = %+v, want item 2", got.Carried.Stack())
	}
}

func TestClickContainerRoundTripsChangedSlots(t *testing.T) {
	original := &v765.ClickContainer{
		WindowID: 0,
		StateID:  12,
		Index:    10,
		Button:   4,
		Mode:     2,
		Changed: codec.Slice[v765.ChangedSlot]{
			{
				Index: 10,
				Item: v765.Slot{
					Present: true,
					Item:    5,
					Count:   32,
				},
			},
			{
				Index: 40,
			},
		},
		Carried: v765.Slot{},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Serverbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("click container not registered")
	}

	got := decoded.(*v765.ClickContainer)
	if !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
