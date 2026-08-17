package pathfinder

import graft "github.com/lrnxzz/graft"

type World interface {
	BlockAt(graft.Position) (graft.BlockState, bool)
}

type Terrain interface {
	Passable(graft.BlockState) bool
	Dangerous(graft.BlockState) bool
	BreakTicks(graft.BlockState, graft.ItemID) (int, bool)
}

type Loadout struct {
	Tool     graft.ItemID
	Digging  bool
	Scaffold int
}

func (l Loadout) building() bool {
	return l.Scaffold > 0
}
