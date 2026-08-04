package gpu

import "github.com/go-gl/gl/v3.3-core/gl"

// Color is a straight RGBA tint in the 0..1 range the shaders work in. It is a
// rendering primitive like Point and Rect, shared by the overlay canvas and the
// world painter, so both speak the same language to a plugin.
type Color struct {
	Red   float32
	Green float32
	Blue  float32
	Alpha float32
}

func RGBA(red, green, blue, alpha float32) Color {
	return Color{
		Red:   red,
		Green: green,
		Blue:  blue,
		Alpha: alpha,
	}
}

func Shade(red, green, blue float32) Color {
	return RGBA(red, green, blue, 1)
}

var White = Shade(1, 1, 1)

// Fade returns the same colour at another opacity, which is how a backdrop or a
// highlight is usually derived from a solid one
func (c Color) Fade(alpha float32) Color {
	c.Alpha = alpha

	return c
}

// Blend is how a source colour carries its own coverage. The atlas keeps the two
// apart; a rendered page hands back colour already multiplied by its alpha, and
// drawing one under the other's rule leaves a dark fringe on every soft edge.
type Blend int

const (
	Straight Blend = iota
	Premultiplied
)

func Blending(mode Blend) {
	if mode == Premultiplied {
		gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

		return
	}

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}
