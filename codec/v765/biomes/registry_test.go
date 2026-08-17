package biomes_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec/v765/biomes"
)

func TestOfResolvesID(t *testing.T) {
	info, ok := biomes.Of(0)
	if !ok {
		t.Fatal("no biome for id 0")
	}
	if info.Name != "badlands" {
		t.Errorf("name = %q, want badlands", info.Name)
	}
}

func TestNamedResolves(t *testing.T) {
	info, ok := biomes.Named("plains")
	if !ok {
		t.Fatal("plains not found")
	}
	if info.Dimension != "overworld" {
		t.Errorf("dimension = %q, want overworld", info.Dimension)
	}
}
