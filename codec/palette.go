package codec

import (
	"fmt"
)

type PaletteType struct {
	Entries     int
	IndirectMax int
	DirectBits  int
}

func (k PaletteType) longs(bitsPerEntry int) int {
	if bitsPerEntry == 0 {
		return 0
	}

	perLong := 64 / bitsPerEntry

	return (k.Entries + perLong - 1) / perLong
}

type PalettedContainer[T ~int32] struct {
	PaletteType
	bitsPerEntry int
	palette      Slice[VarInt]
	data         Slice[Long]
}

func (c PalettedContainer[T]) Len() int {
	return c.Entries
}

func (c PalettedContainer[T]) Get(index int) T {
	if c.bitsPerEntry == 0 {
		if len(c.palette) == 0 {
			return 0
		}

		return T(c.palette[0])
	}

	symbol := c.symbolAt(index)
	if c.palette != nil {
		return T(c.palette[symbol])
	}

	return T(symbol)
}

func (c PalettedContainer[T]) symbolAt(index int) uint64 {
	perLong := 64 / c.bitsPerEntry
	long := c.data[index/perLong].Uint64()

	return long >> uint(index%perLong*c.bitsPerEntry) & (uint64(1)<<c.bitsPerEntry - 1)
}

func (c *PalettedContainer[T]) put(index int, symbol uint64) {
	perLong := 64 / c.bitsPerEntry
	shift := uint(index % perLong * c.bitsPerEntry)
	mask := uint64(1)<<c.bitsPerEntry - 1

	word := c.data[index/perLong].Uint64()
	c.data[index/perLong] = Long(word&^(mask<<shift) | (symbol&mask)<<shift)
}

func (c PalettedContainer[T]) paletteSlot(value T) int {
	for slot, entry := range c.palette {
		if T(entry) == value {
			return slot
		}
	}

	return -1
}

func (c *PalettedContainer[T]) Set(index int, value T) {
	if c.bitsPerEntry == 0 {
		if c.Get(0) == value {
			return
		}
		if len(c.palette) == 0 {
			c.palette = Slice[VarInt]{0}
		}
		c.repack(1)
	}

	if c.palette == nil {
		c.put(index, uint64(value))

		return
	}

	slot := c.paletteSlot(value)
	if slot < 0 {
		if len(c.palette) == 1<<c.bitsPerEntry {
			if c.bitsPerEntry < c.IndirectMax {
				c.repack(c.bitsPerEntry + 1)
			} else {
				c.toDirect()
				c.put(index, uint64(value))

				return
			}
		}

		slot = len(c.palette)
		c.palette = append(c.palette, VarInt(value))
	}

	c.put(index, uint64(slot))
}

func (c *PalettedContainer[T]) repack(bitsPerEntry int) {
	symbols := make([]uint64, c.Entries)
	for index := range symbols {
		if c.bitsPerEntry != 0 {
			symbols[index] = c.symbolAt(index)
		}
	}

	c.bitsPerEntry = bitsPerEntry
	c.data = make(Slice[Long], c.longs(bitsPerEntry))
	for index, symbol := range symbols {
		c.put(index, symbol)
	}
}

func (c *PalettedContainer[T]) toDirect() {
	values := make([]uint64, c.Entries)
	for index := range values {
		values[index] = uint64(c.Get(index))
	}

	c.palette = nil
	c.bitsPerEntry = c.DirectBits
	c.data = make(Slice[Long], c.longs(c.DirectBits))
	for index, value := range values {
		c.put(index, value)
	}
}

func (c *PalettedContainer[T]) Decode(r *Reader) error {
	var bitsPerEntry UByte
	if err := bitsPerEntry.Decode(r); err != nil {
		return err
	}
	c.bitsPerEntry = bitsPerEntry.Int()

	switch {
	case c.bitsPerEntry == 0:
		var value VarInt
		if err := value.Decode(r); err != nil {
			return err
		}
		c.palette = Slice[VarInt]{value}
	case c.bitsPerEntry <= c.IndirectMax:
		if err := c.palette.Decode(r); err != nil {
			return err
		}
	}

	if err := c.data.Decode(r); err != nil {
		return err
	}

	expected := c.longs(c.bitsPerEntry)
	if len(c.data) != expected {
		return r.fail(fmt.Errorf("gocraft: paletted container has %d longs, want %d", len(c.data), expected))
	}

	return nil
}
