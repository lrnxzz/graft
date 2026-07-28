package pathfinder

import gocraft "github.com/lrnxzz/go-craft"

// a climb that tunnels has to clear the landing column and the ceiling the bot
// jumps through, which is the widest any single move gets
const maxBreaks = 3

type corner struct {
	first  gocraft.BlockFace
	second gocraft.BlockFace
}

func cornerBetween(first, second gocraft.BlockFace) corner {
	return corner{
		first:  first,
		second: second,
	}
}

var cardinalFaces = [...]gocraft.BlockFace{
	gocraft.FaceNorth,
	gocraft.FaceSouth,
	gocraft.FaceWest,
	gocraft.FaceEast,
}

var diagonalCorners = [...]corner{
	cornerBetween(gocraft.FaceNorth, gocraft.FaceWest),
	cornerBetween(gocraft.FaceNorth, gocraft.FaceEast),
	cornerBetween(gocraft.FaceSouth, gocraft.FaceWest),
	cornerBetween(gocraft.FaceSouth, gocraft.FaceEast),
}

var leapDirections = [...]gocraft.Position{
	gocraft.At(1, 0, 0),
	gocraft.At(-1, 0, 0),
	gocraft.At(0, 0, 1),
	gocraft.At(0, 0, -1),
	gocraft.At(1, 0, 1),
	gocraft.At(1, 0, -1),
	gocraft.At(-1, 0, 1),
	gocraft.At(-1, 0, -1),
}

type move struct {
	to      gocraft.Position
	cost    float64
	broken  [maxBreaks]gocraft.Position
	breaks  int
	support gocraft.Position
	placing bool
}

func moveTo(to gocraft.Position, cost float64) move {
	return move{
		to:   to,
		cost: cost,
	}
}

func (m *move) breaking(block gocraft.Position, cost float64) {
	m.broken[m.breaks] = block
	m.breaks++
	m.cost += cost
}

func (m *move) building(block gocraft.Position, cost float64) {
	m.support = block
	m.placing = true
	m.cost += cost
}

func (m move) digs() bool {
	return m.breaks > 0
}

func (m move) steps() []Step {
	steps := make([]Step, 0, m.breaks+2)
	for index := range m.breaks {
		steps = append(steps, step(Break, m.broken[index]))
	}
	if m.placing {
		steps = append(steps, step(Place, m.support))
	}

	return append(steps, step(Walk, m.to))
}
