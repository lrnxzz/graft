package graft

type AABB struct {
	Min Vec3d
	Max Vec3d
}

func Box(a, b Vec3d) AABB {
	return AABB{
		Min: Vec3(min(a.X, b.X), min(a.Y, b.Y), min(a.Z, b.Z)),
		Max: Vec3(max(a.X, b.X), max(a.Y, b.Y), max(a.Z, b.Z)),
	}
}

func BoxAround(feet Vec3d, width, height float64) AABB {
	half := width / 2

	return AABB{
		Min: Vec3(feet.X-half, feet.Y, feet.Z-half),
		Max: Vec3(feet.X+half, feet.Y+height, feet.Z+half),
	}
}

func (b AABB) Offset(d Vec3d) AABB {
	return AABB{
		Min: b.Min.Add(d),
		Max: b.Max.Add(d),
	}
}

func (b AABB) OffsetXYZ(dx, dy, dz float64) AABB {
	return b.Offset(Vec3(dx, dy, dz))
}

func (b AABB) Grow(dx, dy, dz float64) AABB {
	return AABB{
		Min: Vec3(b.Min.X-dx, b.Min.Y-dy, b.Min.Z-dz),
		Max: Vec3(b.Max.X+dx, b.Max.Y+dy, b.Max.Z+dz),
	}
}

func (b AABB) Stretch(dx, dy, dz float64) AABB {
	if dx < 0 {
		b.Min.X += dx
	} else {
		b.Max.X += dx
	}
	if dy < 0 {
		b.Min.Y += dy
	} else {
		b.Max.Y += dy
	}
	if dz < 0 {
		b.Min.Z += dz
	} else {
		b.Max.Z += dz
	}

	return b
}

func (b AABB) Center() Vec3d {
	return b.Min.Add(b.Max).Scale(0.5)
}

func (b AABB) Size() Vec3d {
	return b.Max.Sub(b.Min)
}

func (b AABB) Contains(p Vec3d) bool {
	return p.X >= b.Min.X && p.X <= b.Max.X &&
		p.Y >= b.Min.Y && p.Y <= b.Max.Y &&
		p.Z >= b.Min.Z && p.Z <= b.Max.Z
}

func (b AABB) Intersects(o AABB) bool {
	return b.Min.X < o.Max.X && b.Max.X > o.Min.X &&
		b.Min.Y < o.Max.Y && b.Max.Y > o.Min.Y &&
		b.Min.Z < o.Max.Z && b.Max.Z > o.Min.Z
}

func (b AABB) ClampX(o AABB, dx float64) float64 {
	if o.Max.Y <= b.Min.Y || o.Min.Y >= b.Max.Y {
		return dx
	}
	if o.Max.Z <= b.Min.Z || o.Min.Z >= b.Max.Z {
		return dx
	}
	if dx > 0 && o.Max.X <= b.Min.X {
		return min(b.Min.X-o.Max.X, dx)
	}
	if dx < 0 && o.Min.X >= b.Max.X {
		return max(b.Max.X-o.Min.X, dx)
	}

	return dx
}

func (b AABB) ClampY(o AABB, dy float64) float64 {
	if o.Max.X <= b.Min.X || o.Min.X >= b.Max.X {
		return dy
	}
	if o.Max.Z <= b.Min.Z || o.Min.Z >= b.Max.Z {
		return dy
	}
	if dy > 0 && o.Max.Y <= b.Min.Y {
		return min(b.Min.Y-o.Max.Y, dy)
	}
	if dy < 0 && o.Min.Y >= b.Max.Y {
		return max(b.Max.Y-o.Min.Y, dy)
	}

	return dy
}

func (b AABB) ClampZ(o AABB, dz float64) float64 {
	if o.Max.X <= b.Min.X || o.Min.X >= b.Max.X {
		return dz
	}
	if o.Max.Y <= b.Min.Y || o.Min.Y >= b.Max.Y {
		return dz
	}
	if dz > 0 && o.Max.Z <= b.Min.Z {
		return min(b.Min.Z-o.Max.Z, dz)
	}
	if dz < 0 && o.Min.Z >= b.Max.Z {
		return max(b.Max.Z-o.Min.Z, dz)
	}

	return dz
}
