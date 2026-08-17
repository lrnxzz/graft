package agent

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/pathfinder"
	"github.com/lrnxzz/graft/physics"
)

type simTerrain struct{}

func (simTerrain) Passable(state graft.BlockState) bool {
	return state == 0
}

func (simTerrain) Dangerous(state graft.BlockState) bool {
	return state == 9
}

func (simTerrain) BreakTicks(state graft.BlockState, _ graft.ItemID) (int, bool) {
	if state == 0 || state == 9 {
		return 0, false
	}

	return 20, true
}

func simCollider(state graft.BlockState) []graft.AABB {
	if state == 0 {
		return nil
	}

	return []graft.AABB{graft.Box(graft.Vec3(0, 0, 0), graft.Vec3(1, 1, 1))}
}

func walk(world *graft.World, route pathfinder.Route, budget int) (graft.Vec3d, <-chan arrival) {
	start := route.Waypoints()[0]
	player := &graft.Player{Position: start.Center().Offset(0, -0.5, 0)}
	body := physics.New(simCollider)

	var nav navigator
	done := make(chan arrival, 1)
	nav.follow(route, done)

	for range budget {
		command, navigating := nav.tick(world, player)
		if !navigating {
			break
		}

		player.Yaw = command.yaw
		body.Tick(world, player, command.controls)
	}

	return player.Position, done
}

func TestNavigatorWalksFlatGround(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := graft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := graft.At(2, 1, 2)
	goal := pathfinder.GoalAt(graft.At(10, 1, 2))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 600)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}

	gap := end.Horizontal().Distance(graft.Vec2(10.5, 2.5))
	if gap > 1 {
		t.Errorf("ended at %v, want within a block of the goal", end)
	}
}

func TestNavigatorCutsDiagonals(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := graft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := graft.At(2, 1, 2)
	goal := pathfinder.GoalAt(graft.At(10, 1, 10))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 600)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}
}

func TestNavigatorClimbsStairs(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}
	for step := range 3 {
		for x := 5 + step; x < 16; x++ {
			for z := range 16 {
				column.SetBlock(x, 1+step, z, 1)
			}
		}
	}

	world := graft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := graft.At(2, 1, 8)
	goal := pathfinder.GoalAt(graft.At(10, 4, 8))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 1200)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}
}

func TestNavigatorDropsLedges(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
			if x < 8 {
				column.SetBlock(x, 3, z, 1)
			}
		}
	}

	world := graft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := graft.At(2, 4, 8)
	goal := pathfinder.GoalAt(graft.At(12, 1, 8))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 1200)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}
}

func TestNavigatorLeapsGaps(t *testing.T) {
	column := graft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			if x != 5 {
				column.SetBlock(x, 6, z, 1)
			}
		}
	}

	world := graft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := graft.At(2, 7, 8)
	goal := pathfinder.GoalAt(graft.At(8, 7, 8))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 1200)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}
}

func TestNavigatorRoundsWallCorners(t *testing.T) {
	world := graft.NewWorld()
	for cx := int32(0); cx <= 2; cx++ {
		for cz := int32(-1); cz <= 0; cz++ {
			column := graft.ChunkColumn(cx, cz, -64, 384)
			for x := range 16 {
				for z := range 16 {
					column.SetBlock(x, -61, z, 1)
				}
			}
			world.LoadColumn(column)
		}
	}
	for z := -8; z <= 8; z++ {
		for y := -60; y <= -58; y++ {
			world.SetBlock(30, y, z, 1)
		}
	}

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := graft.At(26, -60, 0)
	goal := pathfinder.GoalAt(graft.At(34, -60, 0))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 2000)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}
}
