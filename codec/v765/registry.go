package v765

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
	"github.com/lrnxzz/graft/nbt"
)

type dimensionBounds struct {
	minY   int
	height int
}

var overworld = dimensionBounds{
	minY:   -64,
	height: 384,
}

func (s *Session) onRegistryData(c *codec.Client, p *RegistryData) error {
	registry, ok := nbt.Get[nbt.Compound](nbt.Compound(p.Codec), "minecraft:dimension_type")
	if !ok {
		return nil
	}

	entries, ok := nbt.Get[nbt.List](registry, "value")
	if !ok {
		return nil
	}

	types, ok := nbt.Items[nbt.Compound](entries)
	if !ok {
		return nil
	}

	for _, entry := range types {
		name, ok := nbt.Get[nbt.String](entry, "name")
		if !ok {
			continue
		}

		element, ok := nbt.Get[nbt.Compound](entry, "element")
		if !ok {
			continue
		}

		minY, ok := nbt.Get[nbt.Int](element, "min_y")
		if !ok {
			continue
		}

		height, ok := nbt.Get[nbt.Int](element, "height")
		if !ok {
			continue
		}

		s.dimensions[graft.Identifier(name)] = dimensionBounds{
			minY:   int(minY),
			height: int(height),
		}
	}

	return nil
}
