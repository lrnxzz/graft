package graft

import "math"

type RayHit struct {
	Block    Position
	State    BlockState
	Face     BlockFace
	Point    Vec3d
	Distance float64
}

func (w *World) Raycast(origin, direction Vec3d, reach float64, solid func(BlockState) bool) (RayHit, bool) {
	direction = direction.Normalize()
	if direction.LengthSquared() == 0 || reach <= 0 {
		return RayHit{}, false
	}

	march := func(origin, direction float64) (int, float64, float64) {
		switch {
		case direction > 0:
			return 1, (math.Floor(origin) + 1 - origin) / direction, 1 / direction
		case direction < 0:
			return -1, (origin - math.Floor(origin)) / -direction, -1 / direction
		default:
			return 0, math.Inf(1), math.Inf(1)
		}
	}

	entered := func(step int, positive, negative BlockFace) BlockFace {
		if step > 0 {
			return positive
		}

		return negative
	}

	block := origin.Floor()
	stepX, tMaxX, tDeltaX := march(origin.X, direction.X)
	stepY, tMaxY, tDeltaY := march(origin.Y, direction.Y)
	stepZ, tMaxZ, tDeltaZ := march(origin.Z, direction.Z)

	for {
		var (
			face     BlockFace
			distance float64
		)

		switch {
		case tMaxX <= tMaxY && tMaxX <= tMaxZ:
			distance = tMaxX
			tMaxX += tDeltaX
			block.X += stepX
			face = entered(stepX, FaceWest, FaceEast)
		case tMaxY <= tMaxZ:
			distance = tMaxY
			tMaxY += tDeltaY
			block.Y += stepY
			face = entered(stepY, FaceDown, FaceUp)
		default:
			distance = tMaxZ
			tMaxZ += tDeltaZ
			block.Z += stepZ
			face = entered(stepZ, FaceNorth, FaceSouth)
		}

		if distance > reach {
			return RayHit{}, false
		}

		state, ok := w.BlockAt(block)
		if !ok {
			return RayHit{}, false
		}
		if !solid(state) {
			continue
		}

		return RayHit{
			Block:    block,
			State:    state,
			Face:     face,
			Point:    origin.Add(direction.Scale(distance)),
			Distance: distance,
		}, true
	}
}
