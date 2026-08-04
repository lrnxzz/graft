package physics_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/physics"
)

func TestPhysicsLandsOnGround(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	column.SetBlock(0, 0, 0, 1)

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{Position: gocraft.Vec3(0.5, 5, 0.5)}
	body := physics.New(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{
			gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)),
		}
	})

	for range 200 {
		body.Tick(world, player, gocraft.Controls{})
	}

	if !player.OnGround {
		t.Fatal("player should rest on the ground after falling")
	}
	if got := player.Position.Y; got != 1 {
		t.Errorf("resting Y = %v, want 1", got)
	}
}

func TestPhysicsFallsThroughAir(t *testing.T) {
	world := gocraft.NewWorld()
	world.LoadColumn(gocraft.ChunkColumn(0, 0, -64, 384))

	player := &gocraft.Player{Position: gocraft.Vec3(0.5, 100, 0.5)}
	body := physics.New(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{
			gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)),
		}
	})

	for range 3 {
		body.Tick(world, player, gocraft.Controls{})
	}

	if player.OnGround {
		t.Error("player over empty space should not be on the ground")
	}
	if player.Position.Y >= 100 {
		t.Errorf("player should have fallen, Y = %v", player.Position.Y)
	}
}

func TestPhysicsWalksForward(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{Position: gocraft.Vec3(8, 1, 8), OnGround: true}
	body := physics.New(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{
			gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)),
		}
	})

	for range 20 {
		body.Tick(world, player, gocraft.Controls{Forward: true})
	}

	if player.Position.Z <= 8.5 {
		t.Errorf("walking forward (yaw 0) should advance +Z, got Z = %v", player.Position.Z)
	}
	if !player.OnGround {
		t.Error("player should stay on the ground while walking")
	}
}

func TestPhysicsJumpClearsOneBlock(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{Position: gocraft.Vec3(8, 1, 8), OnGround: true}
	body := physics.New(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{
			gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)),
		}
	})

	peak := player.Position.Y
	for range 12 {
		body.Tick(world, player, gocraft.Controls{Jump: true})
		peak = max(peak, player.Position.Y)
	}

	if peak < 2 {
		t.Errorf("jump should clear a full block, peak Y = %v (want >= 2)", peak)
	}
}

func TestPhysicsStepsOntoSlab(t *testing.T) {
	column := gocraft.ChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}
	column.SetBlock(8, 1, 10, 2)

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{
		Position: gocraft.Vec3(8.5, 1, 8.5),
		OnGround: true,
	}
	body := physics.New(func(state gocraft.BlockState) []gocraft.AABB {
		switch state {
		case 1:
			return []gocraft.AABB{
				gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)),
			}
		case 2:
			return []gocraft.AABB{
				gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 0.5, 1)),
			}
		default:
			return nil
		}
	})

	for range 8 {
		body.Tick(world, player, gocraft.Controls{Forward: true})
	}

	if got := player.Position.Y; got != 1.5 {
		t.Errorf("player Y = %v, want 1.5 standing on the slab", got)
	}
	if z := player.Position.Z; z <= 10 || z >= 11 {
		t.Errorf("player Z = %v, want to be over the slab", z)
	}
}
