package graft

import (
	"fmt"

	"github.com/lrnxzz/graft/codec"
)

const chunkWidth = 16

type ChunkPos struct {
	X int32
	Z int32
}

func Chunk(x, z int32) ChunkPos {
	return ChunkPos{
		X: x,
		Z: z,
	}
}

func (c ChunkPos) Offset(dx, dz int32) ChunkPos {
	return Chunk(c.X+dx, c.Z+dz)
}

// Origin is the block at the north-west corner of the column, the anchor every
// section index is measured from
func (c ChunkPos) Origin() Position {
	return At(int(c.X)*chunkWidth, 0, int(c.Z)*chunkWidth)
}

func (c ChunkPos) String() string {
	return fmt.Sprintf("chunk(%d, %d)", c.X, c.Z)
}

type ChunkSection struct {
	blockCount codec.Short
	blocks     codec.PalettedContainer[BlockState]
	biomes     codec.PalettedContainer[BiomeID]
}

func (s *ChunkSection) Decode(r *codec.Reader) error {
	s.blocks = BlockStates()
	s.biomes = Biomes()

	return codec.DecodeAll(r, &s.blockCount, &s.blocks, &s.biomes)
}

func (s ChunkSection) Empty() bool {
	return s.blockCount == 0
}

func (s ChunkSection) Block(x, y, z int) BlockState {
	return s.blocks.Get(y<<8 | z<<4 | x)
}

func (s *ChunkSection) SetBlock(x, y, z int, state BlockState) {
	index := y<<8 | z<<4 | x

	previous := s.blocks.Get(index)
	if previous == state {
		return
	}

	s.blocks.Set(index, state)
	switch {
	case previous == 0:
		s.blockCount++
	case state == 0:
		s.blockCount--
	}
}

func (s ChunkSection) Biome(x, y, z int) BiomeID {
	return s.biomes.Get(y>>2<<4 | z>>2<<2 | x>>2)
}

type Column struct {
	X        int32
	Z        int32
	minY     int
	revision int
	sections []ChunkSection
}

func ChunkColumn(x, z int32, minY, height int) *Column {
	sections := make([]ChunkSection, height/16)
	for i := range sections {
		sections[i].blocks = BlockStates()
		sections[i].biomes = Biomes()
	}

	return &Column{
		X:        x,
		Z:        z,
		minY:     minY,
		sections: sections,
	}
}

func (c *Column) Pos() ChunkPos {
	return Chunk(c.X, c.Z)
}

func (c *Column) Decode(r *codec.Reader) error {
	for i := range c.sections {
		if err := c.sections[i].Decode(r); err != nil {
			return err
		}
	}

	return nil
}

func (c *Column) Section(index int) *ChunkSection {
	return &c.sections[index]
}

func (c *Column) MinY() int {
	return c.minY
}

func (c *Column) Height() int {
	return len(c.sections) * 16
}

func (c *Column) Block(x, y, z int) BlockState {
	offset := y - c.minY

	return c.sections[offset>>4].Block(x, offset&15, z)
}

func (c *Column) SetBlock(x, y, z int, state BlockState) {
	offset := y - c.minY

	c.revision++
	c.sections[offset>>4].SetBlock(x, offset&15, z, state)
}

func (c *Column) Revision() int {
	return c.revision
}

func (c *Column) Biome(x, y, z int) BiomeID {
	offset := y - c.minY

	return c.sections[offset>>4].Biome(x, offset&15, z)
}
