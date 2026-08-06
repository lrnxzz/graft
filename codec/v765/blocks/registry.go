package blocks

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/assets"
	"github.com/lrnxzz/go-craft/lib"
)

//go:generate go run github.com/lrnxzz/go-craft/cli gen blocks 765

var registry = lib.LoadRegistry[gocraft.Block](765, assets.Blocks)

var Of = lib.Ranged(registry, func(b gocraft.Block) (gocraft.BlockState, gocraft.BlockState) {
	return b.MinState, b.MaxState
})

var Named = lib.Keyed(registry, func(b gocraft.Block) gocraft.Identifier {
	return b.Name
})

var Names = lib.Listed(registry, func(e gocraft.Block) gocraft.Identifier {
	return e.Name
})
