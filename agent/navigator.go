package agent

import (
	"sync"

	"errors"
	"math"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/pathfinder"
)

var (
	errNavigationStuck   = errors.New("agent: navigation stuck")
	errNavigationStopped = errors.New("agent: navigation stopped")
	errNavigationPartial = errors.New("agent: navigation ended before the goal")
)

const (
	waypointRadius  = 0.35
	waypointRise    = 0.3
	leapWindow      = 3.3
	stuckPatience   = 100
	actionPatience  = 300
	stuckEpsilon    = 0.01
	descendMargin   = 0.5
	climbMargin     = 0.3
	leapBlockLength = 2
)

var placeFaces = [...]graft.BlockFace{
	graft.FaceDown,
	graft.FaceNorth,
	graft.FaceSouth,
	graft.FaceWest,
	graft.FaceEast,
	graft.FaceUp,
}

type order struct {
	action   pathfinder.Action
	target   graft.Position
	aim      graft.Vec3d
	controls graft.Controls
	yaw      float32
	pitch    float32
}

// a route can end short of its goal, so unlike a dig the walker has to say where
// it stopped and not just whether it failed; one buffered send keeps the tick
// loop from ever waiting on the caller
type arrival struct {
	at  graft.Position
	err error
}

// the lock lives here rather than on the agent: the tick walks the route while a
// caller may abandon it, and nothing else waits on that.
//
// follow and tick take it and then call what abandon does, so the giving up is
// an unlocked core rather than a second lock the same goroutine would wait on.
type navigator struct {
	mu       sync.Mutex
	steps    []pathfinder.Step
	next     int
	complete bool
	done     chan<- arrival
	closest  float64
	patience int
}

func (n *navigator) follow(route pathfinder.Route, done chan<- arrival) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.give(errNavigationStopped)

	n.steps = route.Steps
	n.next = 0
	n.complete = route.Complete
	n.done = done
	n.closest = math.Inf(1)
	n.patience = stuckPatience
}

func (n *navigator) abandon(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.give(err)
}

func (n *navigator) give(err error) {
	if n.done == nil {
		return
	}

	n.done <- arrival{err: err}
	n.settle()
}

func (n *navigator) arrive(reached graft.Position) {
	err := errNavigationPartial
	if n.complete {
		err = nil
	}

	n.done <- arrival{
		at:  reached,
		err: err,
	}
	n.settle()
}

func (n *navigator) settle() {
	n.steps = nil
	n.done = nil
}

// Path is where the bot is walking and how many of those waypoints are behind it
type Path struct {
	Waypoints []graft.Position
	Walked    int
}

func (n *navigator) route() Path {
	n.mu.Lock()
	defer n.mu.Unlock()

	waypoints := make([]graft.Position, 0, len(n.steps))
	walked := 0

	for index, current := range n.steps {
		if current.Action != pathfinder.Walk {
			continue
		}
		if index < n.next {
			walked++
		}

		waypoints = append(waypoints, current.Target)
	}

	return Path{
		Waypoints: waypoints,
		Walked:    walked,
	}
}

// advance retires the current step and reports whether that finished the route
func (n *navigator) advance(reached graft.Position, patience int) bool {
	n.next++
	n.closest = math.Inf(1)
	n.patience = patience

	if n.next < len(n.steps) {
		return false
	}

	n.arrive(reached)

	return true
}

func (n *navigator) tick(world *graft.World, player *graft.Player) (order, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for n.done != nil {
		current := n.steps[n.next]

		if current.Action != pathfinder.Walk {
			command, done := n.work(world, player, current)
			if !done {
				return command, n.done != nil
			}
			if n.advance(current.Target, actionPatience) {
				return order{}, true
			}

			continue
		}

		command, reached := n.stride(player, current.Target)
		if !reached {
			return command, n.done != nil
		}
		if n.advance(current.Target, stuckPatience) {
			return order{}, true
		}
	}

	return order{}, false
}

// work drives a break or a place: the world itself reports completion, so a
// server that refuses the action simply runs the patience down into a replan
func (n *navigator) work(world *graft.World, player *graft.Player, current pathfinder.Step) (order, bool) {
	if settled(world, current) {
		return order{}, true
	}

	n.patience--
	if n.patience <= 0 {
		n.give(errNavigationStuck)

		return order{}, false
	}

	aim, reachable := aimFor(world, current)
	if !reachable {
		n.give(errNavigationStuck)

		return order{}, false
	}

	yaw, pitch := graft.LookAngles(player.Eye(), aim)

	return order{
		action: current.Action,
		target: current.Target,
		aim:    aim,
		yaw:    yaw,
		pitch:  pitch,
	}, false
}

func settled(world *graft.World, current pathfinder.Step) bool {
	state, known := world.BlockAt(current.Target)
	if !known {
		return false
	}

	if current.Action == pathfinder.Place {
		return blocks.Solid(state)
	}

	return !blocks.Solid(state)
}

// a block is always placed onto the far face of whatever the crosshair hits, so
// the bot aims at the shared face of a neighbour that already exists
func aimFor(world *graft.World, current pathfinder.Step) (graft.Vec3d, bool) {
	if current.Action == pathfinder.Break {
		return current.Target.Center(), true
	}

	for _, face := range placeFaces {
		neighbour := current.Target.Neighbor(face)

		state, known := world.BlockAt(neighbour)
		if !known || !blocks.Solid(state) {
			continue
		}

		toward := face.Opposite().Normal().Scale(0.5)

		return neighbour.Center().Add(toward), true
	}

	return graft.Vec3d{}, false
}

func (n *navigator) stride(player *graft.Player, waypoint graft.Position) (order, bool) {
	feet := player.Position
	toward := waypoint.Center().Horizontal().Sub(feet.Horizontal())
	horizontal := toward.Length()
	vertical := math.Abs(feet.Y - float64(waypoint.Y))

	if horizontal < waypointRadius && vertical < waypointRise {
		return order{}, true
	}

	if !n.progressing(horizontal + vertical) {
		n.give(errNavigationStuck)

		return order{}, false
	}

	yaw := toward.Yaw()
	previous := n.previousWaypoint(waypoint)

	if overshotDrop(previous, waypoint, feet, toward) {
		// past the drop line, reversing at full speed would climb back onto
		// the ledge: airborne we coast and let gravity land us; caught on the
		// very lip of the edge we creep toward the landing column, no sprint
		creeping := order{
			controls: graft.Controls{Forward: player.OnGround},
			yaw:      yaw,
		}

		return creeping, false
	}

	climbing := float64(waypoint.Y) > feet.Y+climbMargin
	leaping := leaps(previous, waypoint) && horizontal <= leapWindow

	striding := order{
		controls: graft.Controls{
			Forward: true,
			Sprint:  true,
			Jump:    climbing || leaping,
		},
		yaw: yaw,
	}

	return striding, false
}

func (n *navigator) previousWaypoint(fallback graft.Position) graft.Position {
	for index := n.next - 1; index >= 0; index-- {
		if n.steps[index].Action == pathfinder.Walk {
			return n.steps[index].Target
		}
	}

	return fallback
}

func (n *navigator) progressing(distance float64) bool {
	if distance < n.closest-stuckEpsilon {
		n.closest = distance
		n.patience = stuckPatience

		return true
	}

	n.patience--

	return n.patience > 0
}

func overshotDrop(previous, waypoint graft.Position, feet graft.Vec3d, toward graft.Vec2d) bool {
	if float64(waypoint.Y) >= feet.Y-descendMargin {
		return false
	}

	segment := waypoint.Horizontal().Sub(previous.Horizontal())

	return toward.Dot(segment) <= 0
}

func leaps(from, to graft.Position) bool {
	dx := to.X - from.X
	dz := to.Z - from.Z

	return to.Y <= from.Y && max(max(dx, -dx), max(dz, -dz)) >= leapBlockLength
}
