package lib

type Heap[T any] struct {
	items []T
	less  func(a, b T) bool
}

func NewHeap[T any](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{less: less}
}

func (h *Heap[T]) Len() int {
	return len(h.items)
}

func (h *Heap[T]) Push(item T) {
	h.items = append(h.items, item)
	h.up(len(h.items) - 1)
}

func (h *Heap[T]) Pop() (T, bool) {
	if len(h.items) == 0 {
		var zero T

		return zero, false
	}

	top := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.items = h.items[:last]
	if last > 0 {
		h.down(0)
	}

	return top, true
}

func (h *Heap[T]) up(child int) {
	for child > 0 {
		parent := (child - 1) / 2
		if !h.less(h.items[child], h.items[parent]) {
			return
		}

		h.items[child], h.items[parent] = h.items[parent], h.items[child]
		child = parent
	}
}

func (h *Heap[T]) down(parent int) {
	for {
		smallest := parent
		left := parent*2 + 1
		right := parent*2 + 2

		if left < len(h.items) && h.less(h.items[left], h.items[smallest]) {
			smallest = left
		}
		if right < len(h.items) && h.less(h.items[right], h.items[smallest]) {
			smallest = right
		}
		if smallest == parent {
			return
		}

		h.items[parent], h.items[smallest] = h.items[smallest], h.items[parent]
		parent = smallest
	}
}
