package pathfinder

import (
	"math"

	gocraft "github.com/lrnxzz/go-craft"
)

type Goal interface {
	Reached(gocraft.Position) bool
	Estimate(gocraft.Position) float64
}

type goalAt struct {
	position gocraft.Position
}

func GoalAt(position gocraft.Position) Goal {
	return goalAt{
		position: position,
	}
}

func (g goalAt) Reached(p gocraft.Position) bool {
	return p == g.position
}

func (g goalAt) Estimate(p gocraft.Position) float64 {
	return octile(p, g.position)
}

type goalNear struct {
	position gocraft.Position
	radius   float64
}

func GoalNear(position gocraft.Position, radius int) Goal {
	return goalNear{
		position: position,
		radius:   float64(radius),
	}
}

func (g goalNear) Reached(p gocraft.Position) bool {
	return octile(p, g.position) <= g.radius
}

func (g goalNear) Estimate(p gocraft.Position) float64 {
	return max(0, octile(p, g.position)-g.radius)
}

// the estimate has to stay optimistic for A* to keep its guarantee, so the
// vertical term uses the cheapest way to cover a block of height — a free fall
func octile(from, to gocraft.Position) float64 {
	dx := math.Abs(float64(from.X - to.X))
	dy := math.Abs(float64(from.Y - to.Y))
	dz := math.Abs(float64(from.Z - to.Z))

	shorter := min(dx, dz)
	longer := max(dx, dz)

	return shorter*costDiagonal + (longer - shorter) + dy*costFall
}
