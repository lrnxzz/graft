package blocks

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/pathfinder"
)

type terrain struct{}

func Terrain() pathfinder.Terrain {
	return terrain{}
}

var hazards = map[graft.Identifier]bool{
	"lava":             true,
	"fire":             true,
	"soul_fire":        true,
	"magma_block":      true,
	"cactus":           true,
	"sweet_berry_bush": true,
	"powder_snow":      true,
	"campfire":         true,
	"soul_campfire":    true,
	"wither_rose":      true,
}

func (terrain) Passable(state graft.BlockState) bool {
	return !Solid(state)
}

func (terrain) Dangerous(state graft.BlockState) bool {
	block, ok := Of(state)
	if !ok {
		return false
	}

	return hazards[block.Name]
}

func (terrain) BreakTicks(state graft.BlockState, held graft.ItemID) (int, bool) {
	return BreakTicks(state, held)
}
