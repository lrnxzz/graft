package gocraft

import "github.com/lrnxzz/go-craft/codec"

type BiomeID int32

var biomePalette = codec.PaletteType{
	Entries:     64,
	IndirectMax: 3,
	DirectBits:  6,
}

func Biomes() codec.PalettedContainer[BiomeID] {
	return codec.PalettedContainer[BiomeID]{
		PaletteType: biomePalette,
	}
}

type Biome struct {
	ID               BiomeID    `json:"id"`
	Name             Identifier `json:"name"`
	Temperature      float32    `json:"temperature"`
	Dimension        Identifier `json:"dimension"`
	HasPrecipitation bool       `json:"has_precipitation"`
}
