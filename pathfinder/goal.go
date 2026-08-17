package pathfinder

import (
	"math"

	graft "github.com/lrnxzz/graft"
)

type Goal interface {
	Reached(graft.Position) bool
	Estimate(graft.Position) float64
}

type goalAt struct {
	position graft.Position
}

func GoalAt(position graft.Position) Goal {
	return goalAt{
		position: position,
	}
}

func (g goalAt) Reached(p graft.Position) bool {
	return p == g.position
}

func (g goalAt) Estimate(p graft.Position) float64 {
	return octile(p, g.position)
}

type goalNear struct {
	position graft.Position
	radius   float64
}

func GoalNear(position graft.Position, radius int) Goal {
	return goalNear{
		position: position,
		radius:   float64(radius),
	}
}

func (g goalNear) Reached(p graft.Position) bool {
	return octile(p, g.position) <= g.radius
}

func (g goalNear) Estimate(p graft.Position) float64 {
	return max(0, octile(p, g.position)-g.radius)
}

// the estimate has to stay optimistic for A* to keep its guarantee, so the
// vertical term uses the cheapest way to cover a block of height — a free fall
func octile(from, to graft.Position) float64 {
	dx := math.Abs(float64(from.X - to.X))
	dy := math.Abs(float64(from.Y - to.Y))
	dz := math.Abs(float64(from.Z - to.Z))

	shorter := min(dx, dz)
	longer := max(dx, dz)

	return shorter*costDiagonal + (longer - shorter) + dy*costFall
}
