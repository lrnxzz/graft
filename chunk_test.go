package graft_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

func singleSection(block, biome codec.VarInt) []byte {
	var payload []byte

	payload = codec.Short(1).Append(payload)

	payload = append(payload, 0)
	payload = codec.AppendVar(payload, block)
	payload = codec.AppendVar(payload, codec.VarInt(0))

	payload = append(payload, 0)
	payload = codec.AppendVar(payload, biome)
	payload = codec.AppendVar(payload, codec.VarInt(0))

	return payload
}

func TestChunkSectionDecodesBlocksAndBiomes(t *testing.T) {
	var section graft.ChunkSection
	if err := section.Decode(codec.NewReader(singleSection(5, 7))); err != nil {
		t.Fatal(err)
	}

	if got := section.Block(1, 2, 3); got != 5 {
		t.Errorf("Block = %d, want 5", got)
	}
	if got := section.Biome(0, 0, 0); got != 7 {
		t.Errorf("Biome = %d, want 7", got)
	}
}

func TestColumnResolvesSectionByWorldY(t *testing.T) {
	payload := append(singleSection(10, 0), singleSection(20, 0)...)

	column := graft.ChunkColumn(3, -4, -64, 32)
	if err := column.Decode(codec.NewReader(payload)); err != nil {
		t.Fatal(err)
	}

	if got := column.Block(1, -60, 2); got != 10 {
		t.Errorf("Block at y=-60 = %d, want 10 (section 0)", got)
	}
	if got := column.Block(1, -40, 2); got != 20 {
		t.Errorf("Block at y=-40 = %d, want 20 (section 1)", got)
	}
}
