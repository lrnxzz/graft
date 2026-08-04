package codec

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type BitSet []uint64

func (b BitSet) Get(i int) bool {
	word := i / 64
	if i < 0 || word >= len(b) {
		return false
	}

	return b[word]>>(uint(i)%64)&1 == 1
}

func (b *BitSet) Set(i int) {
	if i < 0 {
		return
	}

	word := i / 64
	for len(*b) <= word {
		*b = append(*b, 0)
	}

	(*b)[word] |= 1 << (uint(i) % 64)
}

func (b BitSet) Clear(i int) {
	word := i / 64
	if i < 0 || word >= len(b) {
		return
	}

	b[word] &^= 1 << (uint(i) % 64)
}

func (b BitSet) Len() int {
	return len(b) * 64
}

func (b BitSet) Count() int {
	total := 0
	for _, word := range b {
		total += bits.OnesCount64(word)
	}

	return total
}

func (b BitSet) Append(dst []byte) []byte {
	dst = AppendVar(dst, VarInt(len(b)))
	for _, word := range b {
		dst = binary.BigEndian.AppendUint64(dst, word)
	}

	return dst
}

func (b *BitSet) Decode(r *Reader) error {
	var count VarInt
	if err := count.Decode(r); err != nil {
		return err
	}
	if count < 0 || count.Int() > r.Remaining()/8 {
		return r.fail(fmt.Errorf("gocraft: bitset of %d longs is out of range", count))
	}

	words := make(BitSet, 0, min(count.Int(), maxPrealloc))
	for range count.Int() {
		var word Long
		if err := word.Decode(r); err != nil {
			return err
		}

		words = append(words, word.Uint64())
	}

	*b = words

	return nil
}

type FixedBitSet []byte

func NewFixedBitSet(bitCount int) FixedBitSet {
	return make(FixedBitSet, (bitCount+7)/8)
}

func (f FixedBitSet) Get(i int) bool {
	index := i / 8
	if i < 0 || index >= len(f) {
		return false
	}

	return f[index]>>(uint(i)%8)&1 == 1
}

func (f FixedBitSet) Set(i int) {
	index := i / 8
	if i < 0 || index >= len(f) {
		return
	}

	f[index] |= 1 << (uint(i) % 8)
}

func (f FixedBitSet) Clear(i int) {
	index := i / 8
	if i < 0 || index >= len(f) {
		return
	}

	f[index] &^= 1 << (uint(i) % 8)
}

func (f FixedBitSet) Len() int {
	return len(f) * 8
}

func (f FixedBitSet) Count() int {
	total := 0
	for _, b := range f {
		total += bits.OnesCount8(b)
	}

	return total
}

func (f FixedBitSet) Append(dst []byte) []byte {
	return append(dst, f...)
}

func DecodeFixedBitSet(r *Reader, bitCount int) (FixedBitSet, error) {
	raw := r.take((bitCount + 7) / 8)
	if raw == nil {
		return nil, r.Err()
	}

	return FixedBitSet(append([]byte(nil), raw...)), nil
}
