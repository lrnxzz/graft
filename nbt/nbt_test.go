package nbt_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/lrnxzz/graft/nbt"
)

func richTree() nbt.Compound {
	return nbt.Compound{
		"byte":   nbt.Byte(-128),
		"short":  nbt.Short(math.MinInt16),
		"int":    nbt.Int(math.MinInt32),
		"long":   nbt.Long(math.MinInt64),
		"float":  nbt.Float(math.Pi),
		"double": nbt.Double(-math.MaxFloat64),
		"string": nbt.String("graft ⛏"),
		"bytes":  nbt.ByteArray{0, 1, 2, 255},
		"ints":   nbt.IntArray{0, -1, math.MaxInt32},
		"longs":  nbt.LongArray{0, -1, math.MaxInt64},
		"emptyList": nbt.List{
			Elem:  nbt.TagInt,
			Items: []nbt.Tag{},
		},
		"doubles": nbt.List{
			Elem:  nbt.TagDouble,
			Items: []nbt.Tag{nbt.Double(1.5), nbt.Double(2.5)},
		},
		"listOfCompounds": nbt.List{
			Elem: nbt.TagCompound,
			Items: []nbt.Tag{
				nbt.Compound{
					"id": nbt.Int(1),
				},
				nbt.Compound{
					"id": nbt.Int(2),
				},
			},
		},
		"nested": nbt.Compound{
			"empty": nbt.Compound{},
			"flag":  nbt.Byte(1),
		},
	}
}

func TestDecodeRecoversEncodedTree(t *testing.T) {
	original := richTree()

	decoded, err := nbt.Decode(nbt.Encode(original))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("round trip mismatch:\n got %#v\nwant %#v", decoded, original)
	}
}

func TestDecodeNetworkSample(t *testing.T) {
	sample := []byte{
		byte(nbt.TagCompound),
		byte(nbt.TagByte), 0x00, 0x04, 'b', 'y', 't', 'e', 0x7f,
		byte(nbt.TagString), 0x00, 0x03, 's', 't', 'r', 0x00, 0x02, 'h', 'i',
		byte(nbt.TagEnd),
	}

	decoded, err := nbt.Decode(sample)
	if err != nil {
		t.Fatal(err)
	}

	want := nbt.Compound{
		"byte": nbt.Byte(127),
		"str":  nbt.String("hi"),
	}

	if !reflect.DeepEqual(decoded, want) {
		t.Errorf("decoded %#v, want %#v", decoded, want)
	}
}

func TestDecodeRejectsMalformedInput(t *testing.T) {
	tests := []struct {
		input []byte
	}{
		{
			input: nil,
		},
		{
			input: []byte{byte(nbt.TagByte)},
		},
		{
			input: []byte{byte(nbt.TagCompound), byte(nbt.TagInt), 0x00, 0x01, 'x', 0x00},
		},
	}

	for _, tt := range tests {
		if _, err := nbt.Decode(tt.input); err == nil {
			t.Errorf("Decode(%x): expected an error, got nil", tt.input)
		}
	}
}

func TestDecodeNamedRecoversNameAndTree(t *testing.T) {
	root := nbt.Compound{
		"level": nbt.String("overworld"),
		"seed":  nbt.Long(-4096),
	}

	name, decoded, err := nbt.DecodeNamed(nbt.EncodeNamed("Data", root))
	if err != nil {
		t.Fatal(err)
	}

	if name != "Data" {
		t.Errorf("root name = %q, want Data", name)
	}
	if !reflect.DeepEqual(decoded, root) {
		t.Errorf("named round trip mismatch:\n got %#v\nwant %#v", decoded, root)
	}
}

func TestDecodeRejectsDeeplyNestedInput(t *testing.T) {
	bomb := []byte{byte(nbt.TagCompound), byte(nbt.TagList), 0x00, 0x00}

	for range 1000 {
		bomb = append(bomb, byte(nbt.TagList), 0x00, 0x00, 0x00, 0x01)
	}

	if _, err := nbt.Decode(bomb); err == nil {
		t.Error("expected a nesting-depth error, got nil")
	}
}

func TestDecodeNeverPanicsOnTruncatedInput(t *testing.T) {
	full := nbt.Encode(richTree())

	for cut := range len(full) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Decode panicked on a %d-byte prefix: %v", cut, r)
				}
			}()

			nbt.Decode(full[:cut])
		}()
	}
}
