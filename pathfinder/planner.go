package pathfinder

import (
	"slices"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/lib"
)

type Planner struct {
	world   World
	terrain Terrain
	loadout Loadout
}

func NewPlanner(world World, terrain Terrain, loadout Loadout) *Planner {
	return &Planner{
		world:   world,
		terrain: terrain,
		loadout: loadout,
	}
}

type node struct {
	position graft.Position
	parent   *node
	arrival  move
	cost     float64
	priority float64
}

func (p *Planner) Plan(from graft.Position, goal Goal) (Route, bool) {
	if goal.Reached(from) {
		standing := Route{
			Steps:    []Step{step(Walk, from)},
			Complete: true,
		}

		return standing, true
	}

	better := func(a, b *node) bool {
		return a.priority < b.priority
	}
	open := lib.NewHeap(better)

	best := map[graft.Position]float64{}
	best[from] = 0

	start := &node{
		position: from,
		priority: goal.Estimate(from),
	}
	open.Push(start)

	closest := start
	for range searchBudget {
		current, ok := open.Pop()
		if !ok {
			break
		}
		if current.cost > best[current.position] {
			continue
		}

		if goal.Reached(current.position) {
			return assemble(current, true), true
		}
		if goal.Estimate(current.position) < goal.Estimate(closest.position) {
			closest = current
		}

		reach := func(edge move) {
			total := current.cost + edge.cost

			known, seen := best[edge.to]
			if seen && known <= total {
				return
			}

			best[edge.to] = total
			open.Push(&node{
				position: edge.to,
				parent:   current,
				arrival:  edge,
				cost:     total,
				priority: total + goal.Estimate(edge.to),
			})
		}
		p.expand(current.position, reach)
	}

	if closest == start {
		return Route{}, false
	}

	return assemble(closest, false), true
}

func (p *Planner) expand(from graft.Position, reach func(move)) {
	for _, face := range cardinalFaces {
		ahead := from.Neighbor(face)

		p.stepAcross(ahead, reach)
		p.stepUp(from, ahead, reach)
		p.stepDown(ahead, reach)
		p.bridge(from, ahead, reach)
	}

	p.diagonals(from, reach)
	p.leaps(from, reach)
	p.digDown(from, reach)
}

func assemble(target *node, complete bool) Route {
	var chain []*node
	for current := target; current != nil; current = current.parent {
		chain = append(chain, current)
	}
	slices.Reverse(chain)

	steps := make([]Step, 0, len(chain)+1)
	steps = append(steps, step(Walk, chain[0].position))
	for _, current := range chain[1:] {
		steps = append(steps, current.arrival.steps()...)
	}

	return Route{
		Steps:    steps,
		Cost:     target.cost,
		Complete: complete,
	}
}
