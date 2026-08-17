package nbt_test

import (
	"reflect"
	"testing"

	"github.com/lrnxzz/graft/nbt"
)

func TestUnmarshalRecoversMarshaledStruct(t *testing.T) {
	original := registry{
		meta: meta{
			Version: 7,
		},
		Name:    "nether",
		Palette: []string{"netherrack", "soul_sand"},
		Blocks:  []int64{10, 20, 30},
		Heights: []int32{1, 2, 3},
		Dim: dimension{
			Natural:      false,
			Height:       256,
			AmbientLight: 0.1,
			Optional:     "present",
		},
	}

	encoded, err := nbt.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded registry
	if err := nbt.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("round trip mismatch:\n got %#v\nwant %#v", decoded, original)
	}
}

func TestUnmarshalNamedRecoversNameAndValue(t *testing.T) {
	original := meta{
		Version: 42,
	}

	encoded, err := nbt.MarshalNamed("root", original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded meta
	name, err := nbt.UnmarshalNamed(encoded, &decoded)
	if err != nil {
		t.Fatal(err)
	}

	if name != "root" {
		t.Errorf("root name = %q, want root", name)
	}
	if decoded != original {
		t.Errorf("decoded %+v, want %+v", decoded, original)
	}
}

func TestUnmarshalIntoDynamicMap(t *testing.T) {
	encoded, err := nbt.Marshal(map[string]any{
		"seed":  int64(-42),
		"level": "overworld",
	})
	if err != nil {
		t.Fatal(err)
	}

	decoded := map[string]any{}
	if err := nbt.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	if got := decoded["seed"]; got != nbt.Long(-42) {
		t.Errorf("seed = %#v, want nbt.Long(-42)", got)
	}
	if got := decoded["level"]; got != nbt.String("overworld") {
		t.Errorf("level = %#v, want nbt.String(overworld)", got)
	}
}

func TestUnmarshalSkipsUnknownFields(t *testing.T) {
	encoded, err := nbt.Marshal(map[string]any{
		"version": int32(9),
		"unknown": map[string]any{"nested": int64(1)},
		"noise":   []int32{1, 2, 3},
	})
	if err != nil {
		t.Fatal(err)
	}

	var decoded meta
	if err := nbt.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Version != 9 {
		t.Errorf("version = %d, want 9", decoded.Version)
	}
}

func TestUnmarshalRejectsNonPointer(t *testing.T) {
	var target meta

	if err := nbt.Unmarshal(nbt.Encode(nbt.Compound{}), target); err == nil {
		t.Error("expected an error unmarshaling into a non-pointer, got nil")
	}
}

type longArrayHolder struct {
	Values []int64 `nbt:"values"`
}

func TestUnmarshalRejectsOversizedArrayLength(t *testing.T) {
	payload := []byte{
		byte(nbt.TagCompound),
		byte(nbt.TagLongArray),
		0x00, 0x06, 'v', 'a', 'l', 'u', 'e', 's',
		0x7F, 0xFF, 0xFF, 0xFF,
	}

	var target longArrayHolder
	if err := nbt.Unmarshal(payload, &target); err == nil {
		t.Error("expected an error on an array length with no backing data, got nil")
	}
}
