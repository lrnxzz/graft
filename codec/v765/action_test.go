package v765_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
)

func TestPlayerActionRoundTrips(t *testing.T) {
	original := &v765.PlayerAction{
		Status: 2,
		Location: gocraft.Position{
			X: 10,
			Y: 64,
			Z: -3,
		},
		Face:     1,
		Sequence: 5,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Serverbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("player action not registered")
	}

	got := decoded.(*v765.PlayerAction)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestUseItemOnRoundTrips(t *testing.T) {
	original := &v765.UseItemOn{
		Hand: 0,
		Location: gocraft.Position{
			X: 1,
			Y: 70,
			Z: 2,
		},
		Face:     5,
		CursorX:  0.5,
		CursorY:  1,
		CursorZ:  0.25,
		Sequence: 9,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Serverbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("use item on not registered")
	}

	got := decoded.(*v765.UseItemOn)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestAcknowledgeBlockChangeRoundTrips(t *testing.T) {
	original := &v765.AcknowledgeBlockChange{
		Sequence: 12,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(codec.StatePlay, codec.Clientbound, codec.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("acknowledge block change not registered")
	}

	got := decoded.(*v765.AcknowledgeBlockChange)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
