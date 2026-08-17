package codec_test

import (
	"reflect"
	"testing"

	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/nbt"
)

func TestNBTFieldAdvancesReader(t *testing.T) {
	original := codec.NBT{
		"text": nbt.String("hello"),
		"n":    nbt.Int(7),
	}

	payload := codec.AppendAll(nil, original, codec.VarInt(42))

	var (
		got     codec.NBT
		trailer codec.VarInt
	)
	if err := codec.Unmarshal(payload, &got, &trailer); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, original) {
		t.Errorf("nbt round trip got %#v, want %#v", got, original)
	}
	if trailer != 42 {
		t.Errorf("trailer = %d, want 42 (nbt decode did not advance the reader exactly)", trailer)
	}
}

func TestNBTDecodesTagEndAsAbsent(t *testing.T) {
	payload := codec.AppendAll(nil, codec.NBT(nil), codec.VarInt(9))

	var (
		decoded codec.NBT
		trailer codec.VarInt
	)
	if err := codec.Unmarshal(payload, &decoded, &trailer); err != nil {
		t.Fatal(err)
	}

	if decoded != nil {
		t.Errorf("decoded = %v, want nil", decoded)
	}
	if trailer != 9 {
		t.Errorf("trailer = %d, want 9 (end tag must consume exactly one byte)", trailer)
	}
}
