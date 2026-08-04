package mesh

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const vertexPoolCapacity = 8192

var vertexPool = sync.Pool{
	New: func() any {
		buffer := make([]float32, 0, vertexPoolCapacity)

		return &buffer
	},
}

type Geometry struct {
	vertices []float32
	indices  []uint32
}

// a world vertex is a position, a tile corner and a light level
var layout = []gpu.Attribute{
	{
		Location: 0,
		Size:     3,
	},
	{
		Location: 1,
		Size:     2,
	},
	{
		Location: 2,
		Size:     1,
	},
}

func (g Geometry) Upload() *gpu.Mesh {
	mesh := gpu.NewMesh(g.vertices, g.indices,
		layout...)

	recycled := g.vertices[:0]
	vertexPool.Put(&recycled)

	return mesh
}

type builder struct {
	vertices []float32
	quads    int
}

func newBuilder() builder {
	buffer, pooled := vertexPool.Get().(*[]float32)
	if !pooled {
		return builder{}
	}

	return builder{vertices: (*buffer)[:0]}
}

func (b *builder) quad(origin mgl32.Vec3, face cubeFace, uv gpu.UV) {
	texels := [4]mgl32.Vec2{{uv.U0, uv.V1}, {uv.U1, uv.V1}, {uv.U1, uv.V0}, {uv.U0, uv.V0}}
	for i, corner := range face.corners {
		position := origin.Add(corner)
		b.vertices = append(b.vertices,
			position.X(), position.Y(), position.Z(),
			texels[i].X(), texels[i].Y(),
			face.shade)
	}
	b.quads++
}

func (b *builder) geometry() Geometry {
	return Geometry{
		vertices: b.vertices,
		indices:  gpu.QuadIndices(b.quads),
	}
}
