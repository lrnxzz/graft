package blocks

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/pathfinder"
)

type terrain struct{}

func Terrain() pathfinder.Terrain {
	return terrain{}
}

var hazards = map[gocraft.Identifier]bool{
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

func (terrain) Passable(state gocraft.BlockState) bool {
	return !Solid(state)
}

func (terrain) Dangerous(state gocraft.BlockState) bool {
	block, ok := Of(state)
	if !ok {
		return false
	}

	return hazards[block.Name]
}

func (terrain) BreakTicks(state gocraft.BlockState, held gocraft.ItemID) (int, bool) {
	return BreakTicks(state, held)
}
