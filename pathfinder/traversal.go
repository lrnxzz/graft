package pathfinder

import (
	"math"

	graft "github.com/lrnxzz/graft"
)

const (
	maxFall   = 3
	maxLeap   = 4
	leapReach = 4.3
	headroom  = 2
)

type corner struct {
	first  graft.BlockFace
	second graft.BlockFace
}

func cornerBetween(first, second graft.BlockFace) corner {
	return corner{
		first:  first,
		second: second,
	}
}

var cardinalFaces = [...]graft.BlockFace{
	graft.FaceNorth,
	graft.FaceSouth,
	graft.FaceWest,
	graft.FaceEast,
}

var diagonalCorners = [...]corner{
	cornerBetween(graft.FaceNorth, graft.FaceWest),
	cornerBetween(graft.FaceNorth, graft.FaceEast),
	cornerBetween(graft.FaceSouth, graft.FaceWest),
	cornerBetween(graft.FaceSouth, graft.FaceEast),
}

var leapDirections = [...]graft.Position{
	graft.At(1, 0, 0),
	graft.At(-1, 0, 0),
	graft.At(0, 0, 1),
	graft.At(0, 0, -1),
	graft.At(1, 0, 1),
	graft.At(1, 0, -1),
	graft.At(-1, 0, 1),
	graft.At(-1, 0, -1),
}

func (p *Planner) state(at graft.Position) (graft.BlockState, bool) {
	return p.world.BlockAt(at)
}

func (p *Planner) passable(at graft.Position) bool {
	current, known := p.state(at)

	return known && p.terrain.Passable(current) && !p.terrain.Dangerous(current)
}

func (p *Planner) footing(at graft.Position) bool {
	floor, known := p.state(at.Add(0, -1, 0))

	return known && !p.terrain.Passable(floor) && !p.terrain.Dangerous(floor)
}

func (p *Planner) open(at graft.Position) bool {
	return p.passable(at) && p.passable(at.Add(0, 1, 0))
}

func (p *Planner) standable(at graft.Position) bool {
	return p.footing(at) && p.open(at)
}

// clearing reports what it costs to make a block walk-through, counting an
// already empty block as free so callers can treat every column the same way
func (p *Planner) clearing(at graft.Position) (float64, bool) {
	if p.passable(at) {
		return 0, true
	}
	if !p.loadout.Digging {
		return 0, false
	}

	current, known := p.state(at)
	if !known || p.terrain.Dangerous(current) {
		return 0, false
	}

	ticks, breakable := p.terrain.BreakTicks(current, p.loadout.Tool)
	if !breakable {
		return 0, false
	}

	return costTick * float64(ticks), true
}

func (p *Planner) clear(edge *move, at graft.Position) bool {
	cost, reachable := p.clearing(at)
	if !reachable {
		return false
	}
	if cost == 0 && p.passable(at) {
		return true
	}

	edge.breaking(at, cost)

	return true
}

// the head block goes first on purpose: mining from inside a tunnel, the
// crosshair reaches the block at eye level before the one at the feet, so
// clearing the feet first would leave the bot aiming at a block it must not
// break yet
func (p *Planner) clearColumn(edge *move, at graft.Position) bool {
	return p.clear(edge, at.Add(0, 1, 0)) && p.clear(edge, at)
}

func (p *Planner) stepAcross(ahead graft.Position, reach func(move)) {
	if !p.footing(ahead) {
		return
	}

	crossing := moveTo(ahead, costWalk)
	if !p.clearColumn(&crossing, ahead) {
		return
	}

	reach(crossing)
}

func (p *Planner) stepUp(from, ahead graft.Position, reach func(move)) {
	climbed := ahead.Add(0, 1, 0)
	if !p.footing(climbed) {
		return
	}

	rising := moveTo(climbed, costClimb)
	if !p.clear(&rising, from.Add(0, headroom, 0)) {
		return
	}
	if !p.clearColumn(&rising, climbed) {
		return
	}

	reach(rising)
}

func (p *Planner) stepDown(ahead graft.Position, reach func(move)) {
	if p.footing(ahead) {
		return
	}

	dropping := moveTo(ahead, costWalk)
	if !p.clearColumn(&dropping, ahead) {
		return
	}

	for drop := 1; drop <= maxFall; drop++ {
		landing := ahead.Add(0, -drop, 0)

		if p.standable(landing) {
			dropping.to = landing
			dropping.cost += float64(drop) * costFall
			reach(dropping)

			return
		}
		if !p.passable(landing) {
			return
		}
	}
}

func (p *Planner) diagonals(from graft.Position, reach func(move)) {
	for _, diagonal := range diagonalCorners {
		first := from.Neighbor(diagonal.first)
		second := from.Neighbor(diagonal.second)
		across := first.Neighbor(diagonal.second)

		if p.open(first) && p.open(second) && p.standable(across) {
			reach(moveTo(across, costDiagonal))
		}
	}
}

func (p *Planner) leaps(from graft.Position, reach func(move)) {
	if !p.passable(from.Add(0, headroom, 0)) {
		return
	}

	for _, direction := range leapDirections {
		for length := 2; length <= maxLeap; length++ {
			leap := graft.At(direction.X*length, 0, direction.Z*length)

			span := leap.Horizontal().Length()
			if span > leapReach {
				break
			}
			if !p.corridor(from, leap) {
				break
			}

			landing, safe := p.landingBelow(from.Add(leap.X, 0, leap.Z))
			if !safe {
				continue
			}

			reach(moveTo(landing, costParkour+span+float64(from.Y-landing.Y)*costFall))

			break
		}
	}
}

func (p *Planner) corridor(from, leap graft.Position) bool {
	steps := int(leap.Horizontal().Length() * 2)
	visited := from

	for current := 1; current < steps; current++ {
		fraction := float64(current) / float64(steps)
		column := from.Add(scaled(leap.X, fraction), 0, scaled(leap.Z, fraction))
		if column == visited {
			continue
		}

		visited = column
		if !p.open(column) || !p.passable(column.Add(0, headroom, 0)) {
			return false
		}
	}

	return true
}

func (p *Planner) landingBelow(top graft.Position) (graft.Position, bool) {
	for drop := range maxFall + 1 {
		candidate := top.Add(0, -drop, 0)

		if p.standable(candidate) {
			return candidate, true
		}
		if !p.passable(candidate) {
			return graft.Position{}, false
		}
	}

	return graft.Position{}, false
}

func (p *Planner) digDown(from graft.Position, reach func(move)) {
	if !p.loadout.Digging {
		return
	}

	below := from.Add(0, -1, 0)
	if p.passable(below) {
		return
	}

	sinking := moveTo(below, costWalk)
	if !p.clear(&sinking, below) || !sinking.digs() {
		return
	}
	if !p.footing(below) {
		return
	}

	reach(sinking)
}

// bridging places against the block the bot is standing on, which always has a
// free face because every move lands the bot on solid footing — including the
// support a previous bridge step just laid down, which the world does not show
// yet while the route is still being planned
func (p *Planner) bridge(from, ahead graft.Position, reach func(move)) {
	if !p.loadout.building() || p.footing(ahead) {
		return
	}

	support := ahead.Add(0, -1, 0)
	if !p.passable(support) {
		return
	}
	if support.Y != from.Y-1 {
		return
	}

	crossing := moveTo(ahead, costWalk)
	if !p.clearColumn(&crossing, ahead) {
		return
	}
	crossing.building(support, costPlace)

	reach(crossing)
}

func scaled(offset int, fraction float64) int {
	return int(math.Round(float64(offset) * fraction))
}
