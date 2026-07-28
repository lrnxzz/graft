package pathfinder

import gocraft "github.com/lrnxzz/go-craft"

type World interface {
	BlockAt(gocraft.Position) (gocraft.BlockState, bool)
}

type Terrain interface {
	Passable(gocraft.BlockState) bool
	Dangerous(gocraft.BlockState) bool
	BreakTicks(gocraft.BlockState, gocraft.ItemID) (int, bool)
}

type Loadout struct {
	Tool     gocraft.ItemID
	Digging  bool
	Scaffold int
}

func (l Loadout) building() bool {
	return l.Scaffold > 0
}
