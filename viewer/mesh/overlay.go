package mesh

import (
	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

func Overlay(block gocraft.Position, inflate float32, uv gpu.UV) Geometry {
	origin := vec3(float32(block.X)-inflate, float32(block.Y)-inflate, float32(block.Z)-inflate)
	span := 1 + 2*inflate

	b := newBuilder()
	for _, cube := range cubeFaces {
		face := cubeFace{
			corners: stretch(cube.corners, span),
			shade:   1,
		}
		b.quad(origin, face, uv)
	}

	return b.geometry()
}

func stretch(corners [4]mgl32.Vec3, span float32) [4]mgl32.Vec3 {
	var stretched [4]mgl32.Vec3
	for index, corner := range corners {
		stretched[index] = corner.Mul(span)
	}

	return stretched
}
