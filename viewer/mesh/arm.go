package mesh

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lrnxzz/graft/viewer/gpu"
)

const (
	armWidth  = 0.25
	armHeight = 0.75
	armDepth  = 0.25

	skinSize = 64
)

type skinPanel struct {
	corners [4]mgl32.Vec3
	uv      gpu.UV
	shade   float32
}

func skinUV(x0, y0, x1, y1 float32) gpu.UV {
	return gpu.UV{
		U0: x0 / skinSize,
		V0: y0 / skinSize,
		U1: x1 / skinSize,
		V1: y1 / skinSize,
	}
}

// the box is modeled like the skin paints it — shoulder at the top, fist patch
// on the bottom cap (v-flipped, as vanilla maps cube ends); the first-person
// pose then turns the whole arm upside down
var armPanels = [...]skinPanel{
	{
		corners: [4]mgl32.Vec3{
			vec3(0, 0, armDepth),
			vec3(armWidth, 0, armDepth),
			vec3(armWidth, armHeight, armDepth),
			vec3(0, armHeight, armDepth),
		},
		uv:    skinUV(44, 20, 48, 32),
		shade: 0.8,
	},
	{
		corners: [4]mgl32.Vec3{
			vec3(armWidth, 0, 0),
			vec3(0, 0, 0),
			vec3(0, armHeight, 0),
			vec3(armWidth, armHeight, 0),
		},
		uv:    skinUV(52, 20, 56, 32),
		shade: 1,
	},
	{
		corners: [4]mgl32.Vec3{
			vec3(0, 0, 0),
			vec3(0, 0, armDepth),
			vec3(0, armHeight, armDepth),
			vec3(0, armHeight, 0),
		},
		uv:    skinUV(40, 20, 44, 32),
		shade: 0.75,
	},
	{
		corners: [4]mgl32.Vec3{
			vec3(armWidth, 0, armDepth),
			vec3(armWidth, 0, 0),
			vec3(armWidth, armHeight, 0),
			vec3(armWidth, armHeight, armDepth),
		},
		uv:    skinUV(48, 20, 52, 32),
		shade: 0.75,
	},
	{
		corners: [4]mgl32.Vec3{
			vec3(0, armHeight, armDepth),
			vec3(armWidth, armHeight, armDepth),
			vec3(armWidth, armHeight, 0),
			vec3(0, armHeight, 0),
		},
		uv:    skinUV(44, 16, 48, 20),
		shade: 1,
	},
	{
		corners: [4]mgl32.Vec3{
			vec3(0, 0, 0),
			vec3(armWidth, 0, 0),
			vec3(armWidth, 0, armDepth),
			vec3(0, 0, armDepth),
		},
		uv:    skinUV(48, 20, 52, 16),
		shade: 0.9,
	},
}

func Arm() Geometry {
	b := newBuilder()
	for _, panel := range armPanels {
		face := cubeFace{
			corners: panel.corners,
			shade:   panel.shade,
		}
		b.quad(mgl32.Vec3{}, face, panel.uv)
	}

	return b.geometry()
}
