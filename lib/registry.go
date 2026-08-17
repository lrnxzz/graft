package lib

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
)

type Registry[T any] struct {
	version int32
	entries []T
}

// LoadRegistry reads data compiled into the binary, so bad data is a build
// mistake rather than something to hand back at runtime.
func LoadRegistry[T any](version int32, data []byte) *Registry[T] {
	var entries []T
	if err := json.Unmarshal(data, &entries); err != nil {
		panic(fmt.Sprintf("lib: embedded data is invalid: %v", err))
	}

	return &Registry[T]{
		version: version,
		entries: entries,
	}
}

func Keyed[T any, K comparable](r *Registry[T], key func(T) K) func(K) (T, bool) {
	index := make(map[K]T, len(r.entries))
	for _, entry := range r.entries {
		index[key(entry)] = entry
	}

	return func(k K) (T, bool) {
		entry, ok := index[k]

		return entry, ok
	}
}

func Ranged[T any, K cmp.Ordered](r *Registry[T], bounds func(T) (K, K)) func(K) (T, bool) {
	sorted := slices.Clone(r.entries)
	slices.SortFunc(sorted, func(a, b T) int {
		lo, _ := bounds(a)
		other, _ := bounds(b)

		return cmp.Compare(lo, other)
	})

	return func(k K) (T, bool) {
		index := sort.Search(len(sorted), func(i int) bool {
			_, hi := bounds(sorted[i])

			return hi >= k
		})
		if index < len(sorted) {
			lo, hi := bounds(sorted[index])
			if lo <= k && k <= hi {
				return sorted[index], true
			}
		}

		var zero T

		return zero, false
	}
}

// Listed is every key in the registry, sorted. Keyed answers one lookup; this is
// what something offering a choice needs, such as completing a half-typed name.
func Listed[T any, K cmp.Ordered](r *Registry[T], key func(T) K) []K {
	keys := make([]K, 0, len(r.entries))
	for _, entry := range r.entries {
		keys = append(keys, key(entry))
	}

	slices.Sort(keys)

	return slices.Compact(keys)
}
