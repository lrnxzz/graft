package viewer

import (
	_ "embed"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/viewer/gpu"
)

//go:embed assets/shaders/paint.vert
var paintVertexShader string

//go:embed assets/shaders/paint.frag
var paintFragmentShader string

// a world-space line vertex is a position and a tint
var paintLayout = []gpu.Attribute{
	{
		Location: 0,
		Size:     3,
	},
	{
		Location: 1,
		Size:     4,
	},
}

// a block outline sits a hair outside the block so it does not fight the face it
// traces for the same depth
const outlineSwell = 0.002

// Painter is the world-space surface a WorldLayer draws on. Every coordinate it
// takes is in blocks, so a plugin marking up the world never touches a matrix, a
// shader or a vertex buffer.
type Painter struct {
	camera   Camera
	time     float64
	segments []float32
}

// Time is the running clock of the window, what an animated overlay advances on
func (p *Painter) Time() float64 {
	return p.time
}

// Camera lets a layer that brings its own shader reuse the viewer's view and
// projection instead of recomputing them
func (p *Painter) Camera() Camera {
	return p.camera
}

// Line draws a straight segment between two points in the world
func (p *Painter) Line(from, to graft.Vec3d, tint gpu.Color) {
	p.vertex(from, tint)
	p.vertex(to, tint)
}

// Box outlines an axis-aligned volume with its twelve edges
func (p *Painter) Box(box graft.AABB, tint gpu.Color) {
	lo, hi := box.Min, box.Max

	corners := [...]graft.Vec3d{
		graft.Vec3(lo.X, lo.Y, lo.Z),
		graft.Vec3(hi.X, lo.Y, lo.Z),
		graft.Vec3(hi.X, lo.Y, hi.Z),
		graft.Vec3(lo.X, lo.Y, hi.Z),
		graft.Vec3(lo.X, hi.Y, lo.Z),
		graft.Vec3(hi.X, hi.Y, lo.Z),
		graft.Vec3(hi.X, hi.Y, hi.Z),
		graft.Vec3(lo.X, hi.Y, hi.Z),
	}
	for _, edge := range boxEdges {
		p.Line(corners[edge.from], corners[edge.to], tint)
	}
}

// Highlight outlines a single block, the vanilla selection look
func (p *Painter) Highlight(block graft.Position, tint gpu.Color) {
	corner := block.Corner()
	box := graft.Box(corner, corner.Offset(1, 1, 1))

	p.Box(box.Grow(outlineSwell, outlineSwell, outlineSwell), tint)
}

type boxEdge struct {
	from int
	to   int
}

// the bottom ring, the top ring, then the four uprights joining them
var boxEdges = [...]boxEdge{
	{
		from: 0,
		to:   1,
	},
	{
		from: 1,
		to:   2,
	},
	{
		from: 2,
		to:   3,
	},
	{
		from: 3,
		to:   0,
	},
	{
		from: 4,
		to:   5,
	},
	{
		from: 5,
		to:   6,
	},
	{
		from: 6,
		to:   7,
	},
	{
		from: 7,
		to:   4,
	},
	{
		from: 0,
		to:   4,
	},
	{
		from: 1,
		to:   5,
	},
	{
		from: 2,
		to:   6,
	},
	{
		from: 3,
		to:   7,
	},
}

func (p *Painter) vertex(at graft.Vec3d, tint gpu.Color) {
	p.segments = append(p.segments,
		float32(at.X), float32(at.Y), float32(at.Z),
		tint.Red, tint.Green, tint.Blue, tint.Alpha)
}

func (p *Painter) reset(camera Camera, now float64) {
	p.camera = camera
	p.time = now
	p.segments = p.segments[:0]
}

func (p *Painter) paint(program *gpu.Program) {
	if len(p.segments) == 0 {
		return
	}

	program.Use()
	program.Mat4("viewProjection", p.camera.ViewProjection())

	mesh := gpu.NewLines(p.segments,
		paintLayout...)
	mesh.Draw()
	mesh.Delete()
}
