package pathfinder

import (
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
)

type Action uint8

const (
	Walk Action = iota
	Break
	Place
)

func (a Action) String() string {
	switch a {
	case Walk:
		return "walk"
	case Break:
		return "break"
	case Place:
		return "place"
	}

	return fmt.Sprintf("action(%d)", uint8(a))
}

type Step struct {
	Action Action
	Target gocraft.Position
}

func step(action Action, target gocraft.Position) Step {
	return Step{
		Action: action,
		Target: target,
	}
}

type Route struct {
	Steps    []Step
	Cost     float64
	Complete bool
}

func (r Route) Waypoints() []gocraft.Position {
	waypoints := make([]gocraft.Position, 0, len(r.Steps))
	for _, current := range r.Steps {
		if current.Action == Walk {
			waypoints = append(waypoints, current.Target)
		}
	}

	return waypoints
}
