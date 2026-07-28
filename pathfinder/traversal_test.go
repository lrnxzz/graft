package pathfinder_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/pathfinder"
)

func TestPlannerTunnelsThroughTallWalls(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for y := 1; y <= 6; y++ {
		for z := range 16 {
			column.SetBlock(5, y, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	digging := pathfinder.Loadout{Digging: true}
	planner := pathfinder.NewPlanner(world, testTerrain{}, digging)
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a tunnel through the wall", route, ok)
	}

	got := breaks(route)
	want := 2
	if got != want {
		t.Errorf("breaks = %d, want %d: the feet and head blocks of the wall", got, want)
	}

	for _, current := range route.Steps {
		if current.Action == pathfinder.Break && current.Target.X != 5 {
			t.Errorf("broke %v, want only the wall column at x=5", current.Target)
		}
		if current.Action == pathfinder.Walk && current.Target.Y != 1 {
			t.Errorf("waypoint %v climbed the wall, want a tunnel at foot level", current.Target)
		}
	}

	head, feet := breakOrder(route)
	if head < 0 || feet < 0 {
		t.Fatalf("route = %+v, want both wall blocks broken", route)
	}
	if head > feet {
		t.Errorf("broke the feet block at step %d before the head block at %d, "+
			"which leaves the crosshair on a block the route still needs", feet, head)
	}
}

// mining from inside a tunnel the crosshair reaches eye level first, so the
// route has to retire the head block before the one under it
func breakOrder(route pathfinder.Route) (int, int) {
	head, feet := -1, -1
	for index, current := range route.Steps {
		if current.Action != pathfinder.Break {
			continue
		}

		if current.Target.Y == 2 && head < 0 {
			head = index
		}
		if current.Target.Y == 1 && feet < 0 {
			feet = index
		}
	}

	return head, feet
}

// a two block wall is cheaper to climb than to bore through: breaking the top
// block alone leaves a ledge the bot can step up onto
func TestPlannerClimbsShortWallsInsteadOfBoring(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for z := range 16 {
		column.SetBlock(5, 1, z, stone)
		column.SetBlock(5, 2, z, stone)
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	digging := pathfinder.Loadout{Digging: true}
	planner := pathfinder.NewPlanner(world, testTerrain{}, digging)
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	got := breaks(route)
	want := 1
	if got != want {
		t.Errorf("breaks = %d, want %d: only the block capping the wall", got, want)
	}

	climbed := false
	for _, waypoint := range route.Waypoints() {
		if waypoint.X == 5 && waypoint.Y == 2 {
			climbed = true
		}
	}
	if !climbed {
		t.Errorf("waypoints = %v, want the bot to stand on top of the wall", route.Waypoints())
	}
}

func TestPlannerRefusesToTunnelUnbreakableWalls(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for z := range 16 {
		column.SetBlock(5, 1, z, bedrock)
		column.SetBlock(5, 2, z, bedrock)
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	digging := pathfinder.Loadout{Digging: true}
	planner := pathfinder.NewPlanner(world, testTerrain{}, digging)
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 1, 8))

	route, ok := planner.Plan(from, goal)

	if ok && route.Complete {
		t.Errorf("route = %+v, want no complete route through bedrock", route)
	}
}

// mining is priced in walk-equivalent blocks, so a short detour has to win
func TestPlannerWalksAroundWhenTunnellingCostsMore(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for z := range 16 {
		if z == 9 {
			continue
		}

		column.SetBlock(5, 1, z, stone)
		column.SetBlock(5, 2, z, stone)
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	digging := pathfinder.Loadout{Digging: true}
	planner := pathfinder.NewPlanner(world, testTerrain{}, digging)
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	got := breaks(route)
	want := 0
	if got != want {
		t.Errorf("breaks = %d, want %d: the gap at z=9 is cheaper than mining", got, want)
	}
}

func TestPlannerDigsDownToDescend(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for y := range 5 {
			for z := range 16 {
				column.SetBlock(x, y, z, stone)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	digging := pathfinder.Loadout{Digging: true}
	planner := pathfinder.NewPlanner(world, testTerrain{}, digging)
	from := gocraft.At(8, 5, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 2, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a shaft down through the stone", route, ok)
	}
	if breaks(route) == 0 {
		t.Errorf("route = %+v, want the floor blocks broken on the way down", route)
	}

	for _, current := range route.Steps {
		if current.Action == pathfinder.Place {
			t.Errorf("route places %v, want digging only", current.Target)
		}
	}
}

func TestPlannerBridgesGapsWiderThanAJump(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			if x >= 5 && x <= 10 {
				continue
			}

			column.SetBlock(x, 0, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	building := pathfinder.Loadout{Scaffold: 64}
	planner := pathfinder.NewPlanner(world, testTerrain{}, building)
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(13, 1, 8))

	route, ok := planner.Plan(from, goal)

	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a bridge across the chasm", route, ok)
	}
	if places(route) == 0 {
		t.Errorf("route = %+v, want blocks placed into the chasm", route)
	}

	for _, current := range route.Steps {
		if current.Action == pathfinder.Place && current.Target.Y != 0 {
			t.Errorf("placed %v, want the support at the chasm floor level", current.Target)
		}
	}
}

func TestPlannerNeedsScaffoldToBridge(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			if x >= 5 && x <= 10 {
				continue
			}

			column.SetBlock(x, 0, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, testTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(13, 1, 8))

	route, ok := planner.Plan(from, goal)

	if ok && route.Complete {
		t.Errorf("route = %+v, want no complete route with an empty inventory", route)
	}
}

// the ordering that matters: the bot must never be told to walk into a block
// the route has not cleared for it first
func TestPlannerClearsEveryBlockItWalksInto(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, stone)
		}
	}
	for y := 1; y <= 4; y++ {
		for z := range 16 {
			column.SetBlock(5, y, z, stone)
			column.SetBlock(9, y, z, stone)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	digging := pathfinder.Loadout{Digging: true}
	planner := pathfinder.NewPlanner(world, testTerrain{}, digging)
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(12, 1, 8))

	route, ok := planner.Plan(from, goal)
	if !ok || !route.Complete {
		t.Fatalf("route = %+v, ok = %v, want a complete route", route, ok)
	}

	cleared := map[gocraft.Position]bool{}
	for _, current := range route.Steps {
		if current.Action == pathfinder.Break {
			cleared[current.Target] = true

			continue
		}
		if current.Action != pathfinder.Walk {
			continue
		}

		occupied := [...]gocraft.Position{
			current.Target,
			current.Target.Add(0, 1, 0),
		}
		for _, block := range occupied {
			state, known := world.BlockAt(block)
			if known && state == stone && !cleared[block] {
				t.Errorf("walks into %v while it is still solid", block)
			}
		}
	}
}
