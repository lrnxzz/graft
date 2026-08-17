package nbt_test

import (
	"bytes"
	"testing"

	"github.com/lrnxzz/graft/nbt"
)

func TestEncodeCompoundIsDeterministic(t *testing.T) {
	root := nbt.Compound{
		"alpha":   nbt.Int(1),
		"beta":    nbt.String("two"),
		"gamma":   nbt.Long(3),
		"delta":   nbt.Byte(4),
		"epsilon": nbt.Float(5),
	}

	first := nbt.Encode(root)
	for range 20 {
		if got := nbt.Encode(root); !bytes.Equal(got, first) {
			t.Fatalf("Encode is not deterministic:\n  got %x\nfirst %x", got, first)
		}
	}
}
