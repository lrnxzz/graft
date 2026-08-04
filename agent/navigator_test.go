package agent

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/lrnxzz/go-craft/physics"
)

type simTerrain struct{}

func (simTerrain) Passable(state gocraft.BlockState) bool {
	return state == 0
}

func (simTerrain) Dangerous(state gocraft.BlockState) bool {
	return state == 9
}

func (simTerrain) BreakTicks(state gocraft.BlockState, _ gocraft.ItemID) (int, bool) {
	if state == 0 || state == 9 {
		return 0, false
	}

	return 20, true
}

func simCollider(state gocraft.BlockState) []gocraft.AABB {
	if state == 0 {
		return nil
	}

	return []gocraft.AABB{gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1))}
}

func walk(world *gocraft.World, route pathfinder.Route, budget int) (gocraft.Vec3d, <-chan arrival) {
	start := route.Waypoints()[0]
	player := &gocraft.Player{Position: start.Center().Offset(0, -0.5, 0)}
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
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 2)
	goal := pathfinder.GoalAt(gocraft.At(10, 1, 2))

	route, ok := planner.Plan(from, goal)
	if !ok {
		t.Fatal("no route")
	}

	end, done := walk(world, route, 600)

	result := <-done
	if result.err != nil {
		t.Fatalf("navigation error: %v (ended at %v)", result.err, end)
	}

	gap := end.Horizontal().Distance(gocraft.Vec2(10.5, 2.5))
	if gap > 1 {
		t.Errorf("ended at %v, want within a block of the goal", end)
	}
}

func TestNavigatorCutsDiagonals(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 2)
	goal := pathfinder.GoalAt(gocraft.At(10, 1, 10))

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
	column := gocraft.ChunkColumn(0, 0, -64, 384)
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

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 1, 8)
	goal := pathfinder.GoalAt(gocraft.At(10, 4, 8))

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
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
			if x < 8 {
				column.SetBlock(x, 3, z, 1)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 4, 8)
	goal := pathfinder.GoalAt(gocraft.At(12, 1, 8))

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
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			if x != 5 {
				column.SetBlock(x, 6, z, 1)
			}
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	planner := pathfinder.NewPlanner(world, simTerrain{}, pathfinder.Loadout{})
	from := gocraft.At(2, 7, 8)
	goal := pathfinder.GoalAt(gocraft.At(8, 7, 8))

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
	world := gocraft.NewWorld()
	for cx := int32(0); cx <= 2; cx++ {
		for cz := int32(-1); cz <= 0; cz++ {
			column := gocraft.ChunkColumn(cx, cz, -64, 384)
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
	from := gocraft.At(26, -60, 0)
	goal := pathfinder.GoalAt(gocraft.At(34, -60, 0))

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
