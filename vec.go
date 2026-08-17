package graft

import (
	"fmt"
	"math"
)

// Vec2d is the horizontal plane, the only two axes a walking bot steers on.
// Screen space is a different plane and lives with the renderer, so nothing
// here can be mistaken for pixels.
type Vec2d struct {
	X float64
	Z float64
}

func Vec2(x, z float64) Vec2d {
	return Vec2d{
		X: x,
		Z: z,
	}
}

func (v Vec2d) Add(o Vec2d) Vec2d {
	return Vec2(v.X+o.X, v.Z+o.Z)
}

func (v Vec2d) Sub(o Vec2d) Vec2d {
	return Vec2(v.X-o.X, v.Z-o.Z)
}

func (v Vec2d) Scale(s float64) Vec2d {
	return Vec2(v.X*s, v.Z*s)
}

func (v Vec2d) Neg() Vec2d {
	return Vec2(-v.X, -v.Z)
}

func (v Vec2d) Offset(dx, dz float64) Vec2d {
	return Vec2(v.X+dx, v.Z+dz)
}

func (v Vec2d) Dot(o Vec2d) float64 {
	return v.X*o.X + v.Z*o.Z
}

// Cross is the scalar the 2D wedge product leaves behind: its sign tells which
// side of v the other vector falls on
func (v Vec2d) Cross(o Vec2d) float64 {
	return v.X*o.Z - v.Z*o.X
}

func (v Vec2d) LengthSquared() float64 {
	return v.Dot(v)
}

func (v Vec2d) Length() float64 {
	return math.Hypot(v.X, v.Z)
}

func (v Vec2d) DistanceSquared(o Vec2d) float64 {
	return o.Sub(v).LengthSquared()
}

func (v Vec2d) Distance(o Vec2d) float64 {
	return o.Sub(v).Length()
}

func (v Vec2d) Normalize() Vec2d {
	length := v.Length()
	if length == 0 {
		return Vec2d{}
	}

	return v.Scale(1 / length)
}

func (v Vec2d) Lerp(o Vec2d, t float64) Vec2d {
	return Vec2(v.X+(o.X-v.X)*t, v.Z+(o.Z-v.Z)*t)
}

// Yaw is the heading that faces along the vector, in the degrees the protocol
// uses: zero looks toward +Z and the angle grows clockwise
func (v Vec2d) Yaw() float32 {
	return float32(math.Atan2(-v.X, v.Z) * 180 / math.Pi)
}

func (v Vec2d) At(y float64) Vec3d {
	return Vec3(v.X, y, v.Z)
}

func (v Vec2d) ApproxEqual(o Vec2d, epsilon float64) bool {
	return math.Abs(v.X-o.X) <= epsilon &&
		math.Abs(v.Z-o.Z) <= epsilon
}

func (v Vec2d) String() string {
	return fmt.Sprintf("(%.3f, %.3f)", v.X, v.Z)
}

type Vec3d struct {
	X float64
	Y float64
	Z float64
}

func Vec3(x, y, z float64) Vec3d {
	return Vec3d{
		X: x,
		Y: y,
		Z: z,
	}
}

func (v Vec3d) Add(o Vec3d) Vec3d {
	return Vec3(v.X+o.X, v.Y+o.Y, v.Z+o.Z)
}

func (v Vec3d) Sub(o Vec3d) Vec3d {
	return Vec3(v.X-o.X, v.Y-o.Y, v.Z-o.Z)
}

func (v Vec3d) Mul(o Vec3d) Vec3d {
	return Vec3(v.X*o.X, v.Y*o.Y, v.Z*o.Z)
}

func (v Vec3d) Scale(s float64) Vec3d {
	return Vec3(v.X*s, v.Y*s, v.Z*s)
}

func (v Vec3d) Neg() Vec3d {
	return Vec3(-v.X, -v.Y, -v.Z)
}

func (v Vec3d) Offset(dx, dy, dz float64) Vec3d {
	return Vec3(v.X+dx, v.Y+dy, v.Z+dz)
}

func (v Vec3d) Dot(o Vec3d) float64 {
	return v.X*o.X + v.Y*o.Y + v.Z*o.Z
}

func (v Vec3d) Cross(o Vec3d) Vec3d {
	return Vec3d{
		X: v.Y*o.Z - v.Z*o.Y,
		Y: v.Z*o.X - v.X*o.Z,
		Z: v.X*o.Y - v.Y*o.X,
	}
}

func (v Vec3d) LengthSquared() float64 {
	return v.Dot(v)
}

func (v Vec3d) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vec3d) Horizontal() Vec2d {
	return Vec2(v.X, v.Z)
}

func (v Vec3d) HorizontalLength() float64 {
	return v.Horizontal().Length()
}

func (v Vec3d) DistanceSquared(o Vec3d) float64 {
	return o.Sub(v).LengthSquared()
}

func (v Vec3d) Distance(o Vec3d) float64 {
	return o.Sub(v).Length()
}

func (v Vec3d) Normalize() Vec3d {
	length := v.Length()
	if length == 0 {
		return Vec3d{}
	}

	return v.Scale(1 / length)
}

func (v Vec3d) Lerp(o Vec3d, t float64) Vec3d {
	return Vec3d{
		X: v.X + (o.X-v.X)*t,
		Y: v.Y + (o.Y-v.Y)*t,
		Z: v.Z + (o.Z-v.Z)*t,
	}
}

func (v Vec3d) Floor() Position {
	return Position{
		X: int(math.Floor(v.X)),
		Y: int(math.Floor(v.Y)),
		Z: int(math.Floor(v.Z)),
	}
}

func (v Vec3d) ApproxEqual(o Vec3d, epsilon float64) bool {
	return math.Abs(v.X-o.X) <= epsilon &&
		math.Abs(v.Y-o.Y) <= epsilon &&
		math.Abs(v.Z-o.Z) <= epsilon
}

func (v Vec3d) String() string {
	return fmt.Sprintf("(%.3f, %.3f, %.3f)", v.X, v.Y, v.Z)
}
