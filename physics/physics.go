package physics

import (
	"math"

	graft "github.com/lrnxzz/graft"
)

const (
	gravity      = 0.08
	verticalDrag = 0.98
	walkSpeed    = 0.21
	sprintSpeed  = 0.28
	jumpVelocity = 0.42
	stepHeight   = 0.6
)

type Collider func(graft.BlockState) []graft.AABB

type clamper = func(obstacle, moving graft.AABB, delta float64) float64

type Body struct {
	Velocity graft.Vec3d
	collider Collider
}

func New(collider Collider) *Body {
	simulated := Body{
		collider: collider,
	}

	return &simulated
}

func (p *Body) Tick(world *graft.World, player *graft.Player, controls graft.Controls) {
	heading := controls.Heading(player.Yaw)
	speed := walkSpeed
	if controls.Sprint {
		speed = sprintSpeed
	}
	p.Velocity.X = heading.X * speed
	p.Velocity.Z = heading.Z * speed

	if controls.Jump && player.OnGround {
		p.Velocity.Y = jumpVelocity
	}

	box := player.Box()
	moved := p.collide(world, box, p.Velocity)
	if player.OnGround {
		moved = p.stepUp(world, box, p.Velocity, moved)
	}

	player.OnGround = p.Velocity.Y < 0 && moved.Y != p.Velocity.Y
	player.Position = player.Position.Add(moved)

	if moved.X != p.Velocity.X {
		p.Velocity.X = 0
	}
	if moved.Y != p.Velocity.Y {
		p.Velocity.Y = 0
	}
	if moved.Z != p.Velocity.Z {
		p.Velocity.Z = 0
	}

	p.Velocity.Y -= gravity
	p.Velocity.Y *= verticalDrag
}

func (p *Body) stepUp(world *graft.World, box graft.AABB, velocity, moved graft.Vec3d) graft.Vec3d {
	blocked := moved.X != velocity.X || moved.Z != velocity.Z
	if !blocked {
		return moved
	}

	climb := p.collide(world, box, graft.Vec3(velocity.X, stepHeight, velocity.Z))
	if climb.Y <= 0 {
		return moved
	}

	settle := p.collide(world, box.Offset(climb), graft.Vec3(0, -climb.Y, 0))
	stepped := climb.Add(settle)

	if stepped.HorizontalLength() <= moved.HorizontalLength() {
		return moved
	}

	return stepped
}

func (p *Body) collide(world *graft.World, box graft.AABB, velocity graft.Vec3d) graft.Vec3d {
	obstacles := p.obstacles(world, box.Stretch(velocity.X, velocity.Y, velocity.Z))

	slide := func(box graft.AABB, delta float64, clamp clamper) float64 {
		for _, obstacle := range obstacles {
			delta = clamp(obstacle, box, delta)
		}

		return delta
	}

	dy := slide(box, velocity.Y, graft.AABB.ClampY)
	box = box.OffsetXYZ(0, dy, 0)

	// vanilla resolves the larger horizontal axis first; matching it keeps
	// corner slides identical to the server and avoids "moved wrongly" resets
	dx, dz := velocity.X, velocity.Z
	if math.Abs(dx) < math.Abs(dz) {
		dz = slide(box, dz, graft.AABB.ClampZ)
		dx = slide(box.OffsetXYZ(0, 0, dz), dx, graft.AABB.ClampX)

		return graft.Vec3(dx, dy, dz)
	}

	dx = slide(box, dx, graft.AABB.ClampX)
	dz = slide(box.OffsetXYZ(dx, 0, 0), dz, graft.AABB.ClampZ)

	return graft.Vec3(dx, dy, dz)
}

func (p *Body) obstacles(world *graft.World, region graft.AABB) []graft.AABB {
	lo, hi := region.Min.Floor(), region.Max.Floor()

	var boxes []graft.AABB
	for x := lo.X; x <= hi.X; x++ {
		for y := lo.Y; y <= hi.Y; y++ {
			for z := lo.Z; z <= hi.Z; z++ {
				corner := graft.Vec3(float64(x), float64(y), float64(z))

				state, ok := world.Block(x, y, z)
				if !ok {
					boxes = append(boxes, graft.Box(corner, corner.Offset(1, 1, 1)))

					continue
				}

				for _, shape := range p.collider(state) {
					boxes = append(boxes, shape.Offset(corner))
				}
			}
		}
	}

	return boxes
}
