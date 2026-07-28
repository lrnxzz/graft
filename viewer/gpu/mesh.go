package gpu

import "github.com/go-gl/gl/v3.3-core/gl"

type Attribute struct {
	Location uint32
	Size     int32
}

type Mesh struct {
	vao     uint32
	vbo     uint32
	ebo     uint32
	count   int32
	indexed bool
	mode    uint32
}

// NewLines uploads the vertices as unconnected segments, which is what the world
// painter draws its boxes and rays with
func NewLines(vertices []float32, layout ...Attribute) *Mesh {
	mesh := NewMesh(vertices, nil, layout...)
	mesh.mode = gl.LINES

	return mesh
}

func NewMesh(vertices []float32, indices []uint32, layout ...Attribute) *Mesh {
	if len(vertices) == 0 {
		return &Mesh{}
	}

	var stride int32
	for _, attr := range layout {
		stride += attr.Size
	}

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	var offset int32
	for _, attr := range layout {
		gl.EnableVertexAttribArray(attr.Location)
		gl.VertexAttribPointerWithOffset(attr.Location, attr.Size, gl.FLOAT, false, stride*4, uintptr(offset*4))
		offset += attr.Size
	}

	mesh := &Mesh{
		vao:   vao,
		vbo:   vbo,
		count: int32(len(vertices)) / stride,
		mode:  gl.TRIANGLES,
	}
	if len(indices) > 0 {
		gl.GenBuffers(1, &mesh.ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
		mesh.count = int32(len(indices))
		mesh.indexed = true
	}

	return mesh
}

func (m *Mesh) Delete() {
	gl.DeleteVertexArrays(1, &m.vao)
	gl.DeleteBuffers(1, &m.vbo)
	if m.indexed {
		gl.DeleteBuffers(1, &m.ebo)
	}
}

func (m *Mesh) Draw() {
	if m.count == 0 {
		return
	}

	gl.BindVertexArray(m.vao)
	if m.indexed {
		gl.DrawElements(m.mode, m.count, gl.UNSIGNED_INT, nil)

		return
	}

	gl.DrawArrays(m.mode, 0, m.count)
}

func QuadIndices(quads int) []uint32 {
	indices := make([]uint32, 0, quads*6)
	for quad := range uint32(quads) {
		base := quad * 4
		indices = append(indices, base, base+1, base+2, base, base+2, base+3)
	}

	return indices
}
