package plugin

type Anchor string

const (
	TopLeft     Anchor = "top-left"
	Top         Anchor = "top"
	TopRight    Anchor = "top-right"
	Left        Anchor = "left"
	Center      Anchor = "center"
	Right       Anchor = "right"
	BottomLeft  Anchor = "bottom-left"
	Bottom      Anchor = "bottom"
	BottomRight Anchor = "bottom-right"
)

type Node interface {
	Measure(*placement) (float32, float32)
	Paint(*placement, box)
}

type Anchored interface {
	Anchor() Anchor
}

type Color struct {
	Red   float32
	Green float32
	Blue  float32
	Alpha float32
}

func (c Color) Transparent() bool {
	return c.Alpha == 0
}

func (c Color) or(fallback Color) Color {
	if c.Transparent() {
		return fallback
	}

	return c
}

type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Surface interface {
	Size() (width, height float32)
	Fill(x, y, width, height float32, color Color)
	Text(text string, x, y, scale float32, color Color)
	Icon(sprite Sprite, x, y, size float32)
	Measure(text string, scale float32) float32
}
