package biomes

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/assets"
	"github.com/lrnxzz/graft/lib"
)

//go:generate go run github.com/lrnxzz/graft/cli gen biomes 765

var registry = lib.LoadRegistry[graft.Biome](765, assets.Biomes)

var Of = lib.Keyed(registry, func(b graft.Biome) graft.BiomeID {
	return b.ID
})

var Named = lib.Keyed(registry, func(b graft.Biome) graft.Identifier {
	return b.Name
})
