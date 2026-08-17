package blocks

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/assets"
	"github.com/lrnxzz/graft/lib"
)

//go:generate go run github.com/lrnxzz/graft/cli gen blocks 765

var registry = lib.LoadRegistry[graft.Block](765, assets.Blocks)

var Of = lib.Ranged(registry, func(b graft.Block) (graft.BlockState, graft.BlockState) {
	return b.MinState, b.MaxState
})

var Named = lib.Keyed(registry, func(b graft.Block) graft.Identifier {
	return b.Name
})

var Names = lib.Listed(registry, func(e graft.Block) graft.Identifier {
	return e.Name
})
