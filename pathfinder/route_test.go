package pathfinder_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/pathfinder"
)

type actionCase struct {
	action pathfinder.Action
	want   string
}

func TestActionNames(t *testing.T) {
	cases := []actionCase{
		{
			action: pathfinder.Walk,
			want:   "walk",
		},
		{
			action: pathfinder.Break,
			want:   "break",
		},
		{
			action: pathfinder.Place,
			want:   "place",
		},
		{
			action: pathfinder.Action(200),
			want:   "action(200)",
		},
	}

	for _, current := range cases {
		got := current.action.String()
		if got != current.want {
			t.Errorf("Action(%d).String() = %q, want %q", current.action, got, current.want)
		}
	}
}

func TestRouteWaypointsKeepOnlyTheStepsWalked(t *testing.T) {
	route := pathfinder.Route{
		Steps: []pathfinder.Step{
			{
				Action: pathfinder.Walk,
				Target: graft.At(0, 1, 0),
			},
			{
				Action: pathfinder.Break,
				Target: graft.At(1, 1, 0),
			},
			{
				Action: pathfinder.Break,
				Target: graft.At(1, 2, 0),
			},
			{
				Action: pathfinder.Walk,
				Target: graft.At(1, 1, 0),
			},
			{
				Action: pathfinder.Place,
				Target: graft.At(2, 0, 0),
			},
			{
				Action: pathfinder.Walk,
				Target: graft.At(2, 1, 0),
			},
		},
	}

	got := route.Waypoints()
	want := []graft.Position{
		graft.At(0, 1, 0),
		graft.At(1, 1, 0),
		graft.At(2, 1, 0),
	}

	if len(got) != len(want) {
		t.Fatalf("waypoints = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("waypoint %d = %v, want %v", index, got[index], want[index])
		}
	}
}
