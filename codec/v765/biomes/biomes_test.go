package biomes_test

import (
	"testing"

	"github.com/lrnxzz/graft/codec/v765/biomes"
)

func TestGeneratedConstantsMatchRegistry(t *testing.T) {
	plains, ok := biomes.Named("plains")
	if !ok {
		t.Fatal("plains not found")
	}
	if biomes.Plains != plains.ID {
		t.Errorf("Plains = %d, want %d", biomes.Plains, plains.ID)
	}
}
