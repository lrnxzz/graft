package main

import (
	"errors"
	"fmt"
	"strconv"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/pathfinder"
	"github.com/lrnxzz/graft/viewer"
	"github.com/lrnxzz/graft/viewer/gpu"
)

// A route is a path drawn by hand: the player leaves the body, rests the aim on
// a block until it settles, and the bot walks what was drawn. It draws itself, so
// every point is visible from the air while it is being laid out.
type route struct {
	view   *viewer.Viewer
	points []graft.Position
	reach  []reachable
	walked int
	loops  int
}

// a point is unreachable and a point the bot cannot see are different answers,
// and painting both red would be a lie: the bot only knows the chunks it was sent
type reachable int

const (
	unknown reachable = iota
	open
	blocked
)

var routeTint = map[reachable]gpu.Color{
	unknown: gpu.RGBA(0.6, 0.6, 0.65, 0.8),
	open:    gpu.RGBA(0.49, 0.83, 0.99, 0.9),
	blocked: gpu.RGBA(0.97, 0.44, 0.44, 0.95),
}

func (r *route) mark(at graft.Position) {
	r.points = append(r.points, at)
	r.judge()

	viewer.Says(r.view.Chat()).Row(len(r.points), fmt.Sprintf("%d %d %d", at.X, at.Y, at.Z), spoken(r.reach[len(r.reach)-1]))
}

// judge asks the planner whether each leg can be walked, which is the only
// honest way to know before the bot tries
func (r *route) judge() {
	r.reach = make([]reachable, len(r.points))

	bot := r.view.Bot()
	world := bot.World()

	carrying := pathfinder.Loadout{
		Tool:    bot.Inventory().Held().Item,
		Digging: false,
	}

	planner := pathfinder.NewPlanner(world, blocks.Terrain(), carrying)

	for index := range r.points {
		if index == 0 {
			r.reach[index] = open

			continue
		}

		from, to := r.points[index-1], r.points[index]

		if !world.Surrounds(to, 1) {
			r.reach[index] = unknown

			continue
		}

		if _, found := planner.Plan(from, pathfinder.GoalAt(to)); found {
			r.reach[index] = open

			continue
		}

		r.reach[index] = blocked
	}
}

// DrawWorld paints the route in the world: a line between the points and a stack
// over each, coloured by whether the bot can get there
func (r *route) DrawWorld(painter *viewer.Painter) {
	for index, at := range r.points {
		tint := routeTint[r.reach[index]]

		painter.Highlight(at, tint)
		painter.Line(at.Center(), at.Center().Offset(0, 3, 0), tint)

		if index == 0 {
			continue
		}

		painter.Line(r.points[index-1].Center(), at.Center(), tint)
	}
}

func (r *route) clear() {
	r.points = nil
	r.reach = nil
	r.walked = 0
}

var errNoRoute = errors.New("no route has been drawn — press F, rest your aim on a block for three seconds")

func spoken(state reachable) string {
	switch state {
	case open:
		return viewer.Green + "reachable"
	case blocked:
		return viewer.Red + "no path"
	default:
		return viewer.Faint + "out of sight"
	}
}

func (r *route) describe() string {
	return strconv.Itoa(len(r.points)) + " points"
}
