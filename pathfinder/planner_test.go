package pathfinder_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/pathfinder"
)

const (
	air        gocraft.BlockState = 0
	stone      gocraft.BlockState = 1
	bedrock    gocraft.BlockState = 2
	lava       gocraft.BlockState = 9
	stoneTicks                    = 20
)

type testTerrain struct{}

func (testTerrain) Passable(state gocraft.BlockState) bool {
	return state == air
}

func (testTerrain) Dangerous(state gocraft.BlockState) bool {
	return state == lava
}

func (testTerrain) BreakTicks(state gocraft.BlockState, _ gocraft.ItemID) (int, bool) {
	if state != stone {
		return 0, false
	}

	return stoneTicks, true
}

func breaks(route pathfinder.Route) int {
	total := 0
	for _, current := range route.Steps {
		if current.Action == pathfinder.Break {
			total++
		}
	}

	return total
}

func places(route pathfinder.Route) int {
	total := 0
	for _, current := range route.Steps {
		if current.Action == pathfinder.Place {
			total++
		}
	}

	return total
}

func TestPlannerWalksStraightOnFlatGround(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 2)
	goal := pathfinder.GoalAt(gocraft.At(10, 1, 2))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	got := len(route.Waypoints())
	want := 9
	if got != want {
		t.Errorf("waypoints = %d, want %d on a straight line", got, want)
	}
	if route.Cost != 8 {
		t.Errorf("cost = %v, want 8", route.Cost)
	}
}

func TestPlannerCutsDiagonals(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 2)
	goal := pathfinder.GoalAt(gocraft.At(6, 1, 6))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	got := len(route.Waypoints())
	want := 5
	if got != want {
		t.Errorf("waypoints = %d, want %d pure diagonals", got, want)
	}
	if route.Cost >= 8 {
		t.Errorf("cost = %v, want cheaper than the cardinal route", route.Cost)
	}
}

func TestPlannerRespectsCorners(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for y := 1; y <= 2; y++ {
		column.SetBlock(3, y, 2, stone)
		column.SetBlock(2, y, 3, stone)
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 2)
	goal := pathfinder.GoalAt(gocraft.At(3, 1, 3))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}
	if route.Cost <= 2 {
		t.Errorf("cost = %v, want a detour instead of cutting the blocked corner", route.Cost)
	}
}

func TestPlannerDetoursAroundWalls(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for z := range 15 {
		column.SetBlock(5, 1, z, stone)
		column.SetBlock(5, 2, z, stone)
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	crossed := false
	for _, waypoint := range route.Waypoints() {
		if waypoint.X == 5 && waypoint.Z >= 14 {
			crossed = true
		}
	}
	if !crossed {
		t.Errorf("waypoints = %v, want the detour through the wall gap", route.Waypoints())
	}
}

func TestPlannerClimbsStairs(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for step := range 3 {
		for x := 5 + step; x < 16; x++ {
			for z := range 16 {
				column.SetBlock(x, 1+step, z, stone)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(10, 4, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route up the stairs", route, ok)
	}
}

func TestPlannerDropsSafeLedges(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
			if x < 8 {
				column.SetBlock(x, 3, z, stone)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 4, 8)
	goal := pathfinder.GoalAt(gocraft.At(12, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route off the ledge", route, ok)
	}
}

func TestPlannerRefusesDeadlyDrops(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
			if x < 8 {
				column.SetBlock(x, 5, z, stone)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 6, 8)
	goal := pathfinder.GoalAt(gocraft.At(12, 1, 8))

	route, ok := planner.Plan(from, goal)

	if ok && route.Complete {
		t.Errorf("route = %+v, want no complete route over a five block fall", route)
	}
}

func TestPlannerAvoidsHazards(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for z := range 15 {
		column.SetBlock(5, 0, z, lava)
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route around the lava", route, ok)
	}

	for _, waypoint := range route.Waypoints() {
		if waypoint.X == 5 && waypoint.Z < 15 {
			t.Fatalf("waypoint %v stands on lava", waypoint)
		}
	}
}

func TestPlannerParkoursGaps(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			if x != 5 {
				column.SetBlock(x, 6, z, stone)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 7, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 7, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a parkour jump across the gap", route, ok)
	}

	for _, waypoint := range route.Waypoints() {
		if waypoint.X == 5 {
			t.Fatalf("waypoint %v landed inside the gap", waypoint)
		}
	}
}

func TestPlannerReturnsPartialRoutes(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for y := 1; y <= 3; y++ {
		for offset := -1; offset <= 1; offset++ {
			column.SetBlock(12+offset, y, 8-1, bedrock)
			column.SetBlock(12+offset, y, 8+1, bedrock)
			column.SetBlock(12-1, y, 8, bedrock)
			column.SetBlock(12+1, y, 8, bedrock)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(12, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok {
		t.Fatal("want a best effort route toward the sealed goal")
	}
	if route.Complete {
		t.Fatalf("route = %+v, want an incomplete route", route)
	}

	waypoints := route.Waypoints()
	last := waypoints[len(waypoints)-1]
	if last.X <= 2 {
		t.Errorf("last waypoint = %v, want progress toward the goal", last)
	}
}

func TestPlannerReachesNearGoals(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 2)
	goal := pathfinder.GoalNear(gocraft.At(10, 1, 2), 2)

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	waypoints := route.Waypoints()
	last := waypoints[len(waypoints)-1]
	if last.X < 8 {
		t.Errorf("last waypoint = %v, want within two blocks of the goal", last)
	}
}
