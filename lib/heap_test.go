package lib_test

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/lrnxzz/graft/lib"
)

func TestHeapPopsInOrder(t *testing.T) {
	ascending := func(a, b int) bool {
		return a < b
	}

	heap := lib.NewHeap(ascending)
	values := rand.New(rand.NewSource(7)).Perm(200)
	for _, value := range values {
		heap.Push(value)
	}

	popped := make([]int, 0, len(values))
	for heap.Len() > 0 {
		value, ok := heap.Pop()
		if !ok {
			t.Fatal("pop reported empty with items left")
		}
		popped = append(popped, value)
	}

	if !slices.IsSorted(popped) {
		t.Errorf("popped out of order: %v", popped[:10])
	}
	if len(popped) != len(values) {
		t.Errorf("popped %d values, want %d", len(popped), len(values))
	}
}

func TestHeapPopOnEmpty(t *testing.T) {
	ascending := func(a, b int) bool {
		return a < b
	}

	heap := lib.NewHeap(ascending)
	if _, ok := heap.Pop(); ok {
		t.Error("empty heap should report no value")
	}
}
