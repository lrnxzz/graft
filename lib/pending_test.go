package lib_test

import (
	"testing"

	"github.com/lrnxzz/graft/lib"
)

func TestPendingSequencesAreMonotonic(t *testing.T) {
	var pending lib.Pending[string]

	first := pending.Push("a")
	second := pending.Push("b")
	if first != 1 || second != 2 {
		t.Errorf("sequences = %d, %d, want 1, 2", first, second)
	}
	if pending.Len() != 2 {
		t.Errorf("len = %d, want 2", pending.Len())
	}
}

func TestPendingAckSettlesCumulatively(t *testing.T) {
	var pending lib.Pending[string]

	pending.Push("a")
	second := pending.Push("b")
	pending.Push("c")

	settled := pending.Ack(second)
	if len(settled) != 2 || settled[0] != "a" || settled[1] != "b" {
		t.Errorf("settled = %v, want [a b]", settled)
	}
	if pending.Len() != 1 {
		t.Errorf("len after ack = %d, want 1", pending.Len())
	}
}

func TestPendingAckBeforeAnyPush(t *testing.T) {
	var pending lib.Pending[int]

	if settled := pending.Ack(5); settled != nil {
		t.Errorf("settled = %v, want nil", settled)
	}
}

func TestPendingAckKeepsLaterEntries(t *testing.T) {
	var pending lib.Pending[int]

	pending.Push(10)
	pending.Ack(1)

	next := pending.Push(20)
	if next != 2 {
		t.Errorf("sequence after ack = %d, want 2 (numbering never restarts)", next)
	}
}
