package codec_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

func TestPalettedContainerDecodesSingleValued(t *testing.T) {
	payload := []byte{0}
	payload = codec.AppendVar(payload, codec.VarInt(42))
	payload = codec.AppendVar(payload, codec.VarInt(0))

	container := graft.BlockStates()
	if err := container.Decode(codec.NewReader(payload)); err != nil {
		t.Fatal(err)
	}

	for _, index := range []int{0, 100, 4095} {
		if got := container.Get(index); got != 42 {
			t.Errorf("Get(%d) = %d, want 42", index, got)
		}
	}
}

func TestPalettedContainerDecodesIndirect(t *testing.T) {
	payload := []byte{4}

	payload = codec.AppendVar(payload, codec.VarInt(3))
	for _, value := range []codec.VarInt{10, 20, 30} {
		payload = codec.AppendVar(payload, value)
	}

	longs := make([]uint64, 256)
	longs[0] = 0<<0 | 1<<4 | 2<<8

	payload = codec.AppendVar(payload, codec.VarInt(len(longs)))
	for _, long := range longs {
		payload = codec.Long(long).Append(payload)
	}

	container := graft.BlockStates()
	if err := container.Decode(codec.NewReader(payload)); err != nil {
		t.Fatal(err)
	}

	want := map[int]graft.BlockState{0: 10, 1: 20, 2: 30, 3: 10, 4095: 10}
	for index, state := range want {
		if got := container.Get(index); got != state {
			t.Errorf("Get(%d) = %d, want %d", index, got, state)
		}
	}
}

func TestPalettedContainerSetUpgradesFromSingle(t *testing.T) {
	container := graft.BlockStates()

	container.Set(0, 5)
	container.Set(1, 5)
	container.Set(2, 9)

	want := map[int]graft.BlockState{0: 5, 1: 5, 2: 9, 3: 0, 4095: 0}
	for index, state := range want {
		if got := container.Get(index); got != state {
			t.Errorf("Get(%d) = %d, want %d", index, got, state)
		}
	}
}

func TestPalettedContainerSetOverflowsToDirect(t *testing.T) {
	container := graft.Biomes()

	for i := 0; i < 20; i++ {
		container.Set(i, graft.BiomeID(i))
	}

	for i := 0; i < 20; i++ {
		if got := container.Get(i); got != graft.BiomeID(i) {
			t.Errorf("Get(%d) = %d, want %d", i, got, i)
		}
	}
}
