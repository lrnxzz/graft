package pathfinder

import graft "github.com/lrnxzz/graft"

// a climb that tunnels has to clear the landing column and the ceiling the bot
// jumps through, which is the widest any single move gets
const maxBreaks = 3

type move struct {
	to      graft.Position
	cost    float64
	broken  [maxBreaks]graft.Position
	breaks  int
	support graft.Position
	placing bool
}

func moveTo(to graft.Position, cost float64) move {
	return move{
		to:   to,
		cost: cost,
	}
}

func (m *move) breaking(block graft.Position, cost float64) {
	m.broken[m.breaks] = block
	m.breaks++
	m.cost += cost
}

func (m *move) building(block graft.Position, cost float64) {
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
