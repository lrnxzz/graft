package codec_test

import (
	"reflect"
	"testing"

	"github.com/lrnxzz/go-craft/codec"
)

func TestBitSetSetGetClear(t *testing.T) {
	var set codec.BitSet

	indices := []int{0, 1, 63, 64, 130, 200}
	for _, i := range indices {
		set.Set(i)
	}

	for _, i := range indices {
		if !set.Get(i) {
			t.Errorf("Get(%d) = false after Set, want true", i)
		}
	}

	if got := set.Count(); got != len(indices) {
		t.Errorf("Count() = %d, want %d", got, len(indices))
	}
	if set.Get(5) {
		t.Error("Get(5) = true, want false (never set)")
	}

	set.Clear(64)
	if set.Get(64) {
		t.Error("Get(64) = true after Clear, want false")
	}
	if got := set.Count(); got != len(indices)-1 {
		t.Errorf("Count() after Clear = %d, want %d", got, len(indices)-1)
	}
}

func TestBitSetIgnoresNegativeIndex(t *testing.T) {
	var set codec.BitSet

	set.Set(-1)
	set.Set(-30)

	if set.Get(-1) || set.Get(-30) {
		t.Error("Get(negative) = true, want false")
	}
	if got := set.Count(); got != 0 {
		t.Errorf("Count() = %d after negative Set, want 0", got)
	}

	fixed := codec.NewFixedBitSet(16)
	fixed.Set(-1)
	fixed.Set(-5)

	if fixed.Get(-1) || fixed.Count() != 0 {
		t.Error("FixedBitSet corrupted by a negative index")
	}
}

func TestBitSetRecoversSetBits(t *testing.T) {
	var want codec.BitSet
	for _, i := range []int{3, 70, 128, 255} {
		want.Set(i)
	}

	var got codec.BitSet
	if err := codec.Unmarshal(want.Append(nil), &got); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip yielded %v, want %v", got, want)
	}
}

func TestFixedBitSetRecoversSetBits(t *testing.T) {
	const bitCount = 26

	want := codec.NewFixedBitSet(bitCount)
	for _, i := range []int{0, 7, 8, 25} {
		want.Set(i)
	}

	got, err := codec.DecodeFixedBitSet(codec.NewReader(want.Append(nil)), bitCount)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip yielded %v, want %v", got, want)
	}
	if got.Count() != 4 {
		t.Errorf("Count() = %d, want 4", got.Count())
	}
	if got.Len() < bitCount {
		t.Errorf("Len() = %d, want at least %d", got.Len(), bitCount)
	}
}
