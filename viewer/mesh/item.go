package mesh

import (
	"github.com/go-gl/mathgl/mgl32"
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/viewer/gpu"
)

var spriteFace = cubeFace{
	face: Side,
	corners: [4]mgl32.Vec3{
		vec3(0, 0, 0),
		vec3(1, 0, 0),
		vec3(1, 1, 0),
		vec3(0, 1, 0),
	},
	shade: 1.0,
}

func Block(state graft.BlockState, blocks Blocks) Geometry {
	b := newBuilder()
	for _, face := range cubeFaces {
		b.quad(mgl32.Vec3{}, face, blocks.Tile(state, face.face))
	}

	return b.geometry()
}

func Sprite(uv gpu.UV) Geometry {
	b := newBuilder()
	b.quad(mgl32.Vec3{}, spriteFace, uv)

	return b.geometry()
}
