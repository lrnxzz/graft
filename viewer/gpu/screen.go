package gpu

// Point is a spot in screen space, in pixels measured from the top-left corner.
// It is deliberately float32: everything downstream of it is vertex data, so
// keeping the window in the meshes' own scalar spares every call site a cast.
type Point struct {
	X float32
	Y float32
}

func At(x, y float32) Point {
	return Point{
		X: x,
		Y: y,
	}
}

func (p Point) Offset(dx, dy float32) Point {
	return At(p.X+dx, p.Y+dy)
}

func (p Point) Scale(s float32) Point {
	return At(p.X*s, p.Y*s)
}

func (p Point) Add(o Point) Point {
	return At(p.X+o.X, p.Y+o.Y)
}

func (p Point) Sub(o Point) Point {
	return At(p.X-o.X, p.Y-o.Y)
}

type Rect struct {
	Min Point
	Max Point
}

func RectAt(origin Point, width, height float32) Rect {
	return Rect{
		Min: origin,
		Max: origin.Offset(width, height),
	}
}

func (r Rect) Width() float32 {
	return r.Max.X - r.Min.X
}

func (r Rect) Height() float32 {
	return r.Max.Y - r.Min.Y
}

func (r Rect) Center() Point {
	return At((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
}

// Centered fits a rectangle of that size in the middle of this one, which is
// where the vanilla HUD anchors the crosshair and the container panel
func (r Rect) Centered(width, height float32) Rect {
	origin := r.Center().Offset(-width/2, -height/2)

	return RectAt(origin, width, height)
}

func (r Rect) Offset(dx, dy float32) Rect {
	return Rect{
		Min: r.Min.Offset(dx, dy),
		Max: r.Max.Offset(dx, dy),
	}
}

func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X < r.Max.X &&
		p.Y >= r.Min.Y && p.Y < r.Max.Y
}
