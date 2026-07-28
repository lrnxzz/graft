package viewer

import (
	_ "embed"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

//go:embed assets/shaders/route.vert
var routeVertexShader string

//go:embed assets/shaders/route.frag
var routeFragmentShader string

const (
	routeLift       = 0.06
	routeHalfWidth  = 0.07
	beaconHeight    = 2.4
	beaconHalfWidth = 0.16

	routeFloatsPerQuad = 16
	degenerateSegment  = 1e-3
)

var (
	routeUp     = mgl32.Vec3{0, 1, 0}
	beaconLift  = mgl32.Vec3{0, beaconHeight, 0}
	beaconSides = [...]mgl32.Vec3{
		{beaconHalfWidth, 0, 0},
		{0, 0, beaconHalfWidth},
	}
)

// Route is a WorldLayer that brings its own shader: the dashed ribbon cannot be
// expressed as painter lines, which is exactly why a layer is free to draw its
// own geometry instead of being forced through the painter
type Route struct {
	bot     *agent.Agent
	program *gpu.Program
	mesh    *gpu.Mesh
	points  []mgl32.Vec3
	marks   []float32
	first   gocraft.Position
	last    gocraft.Position
	walked  float32
}

func NewRoute(bot *agent.Agent) (*Route, error) {
	program, err := gpu.NewProgram(routeVertexShader, routeFragmentShader)
	if err != nil {
		return nil, err
	}

	return &Route{
		bot:     bot,
		program: program,
	}, nil
}

func (r *Route) DrawWorld(painter *Painter) {
	waypoints, next := r.bot.Route()
	r.update(waypoints, next, r.bot.Snapshot().Position)

	if r.mesh == nil {
		return
	}

	r.program.Use()
	r.program.Mat4("viewProjection", painter.Camera().ViewProjection())
	r.program.Float("walked", r.walked)
	r.program.Float("time", float32(painter.Time()))
	r.mesh.Draw()
}

func (r *Route) update(waypoints []gocraft.Position, next int, feet gocraft.Vec3d) {
	if len(waypoints) < 2 {
		r.clear()

		return
	}

	if len(waypoints) != len(r.points) || waypoints[0] != r.first || waypoints[len(waypoints)-1] != r.last {
		r.rebuild(waypoints)
	}

	r.walked = r.progress(next, feet)
}

func (r *Route) Close() {
	r.clear()
	r.program.Delete()
}

func (r *Route) rebuild(waypoints []gocraft.Position) {
	r.clear()
	r.first = waypoints[0]
	r.last = waypoints[len(waypoints)-1]

	r.points = make([]mgl32.Vec3, len(waypoints))
	r.marks = make([]float32, len(waypoints))
	for index, waypoint := range waypoints {
		center := waypoint.Center()
		r.points[index] = mgl32.Vec3{
			float32(center.X),
			float32(waypoint.Y) + routeLift,
			float32(center.Z),
		}
		if index > 0 {
			r.marks[index] = r.marks[index-1] + r.points[index].Sub(r.points[index-1]).Len()
		}
	}

	vertices := make([]float32, 0, (len(waypoints)+1)*routeFloatsPerQuad)
	emit := func(point mgl32.Vec3, along float32) {
		vertices = append(vertices, point.X(), point.Y(), point.Z(), along)
	}

	for index := range r.points[:len(r.points)-1] {
		from := r.points[index]
		to := r.points[index+1]

		side := to.Sub(from).Cross(routeUp)
		if side.Len() < degenerateSegment {
			side = mgl32.Vec3{1, 0, 0}
		}
		side = side.Normalize().Mul(routeHalfWidth)

		emit(from.Sub(side), r.marks[index])
		emit(from.Add(side), r.marks[index])
		emit(to.Add(side), r.marks[index+1])
		emit(to.Sub(side), r.marks[index+1])
	}

	goal := r.points[len(r.points)-1]
	total := r.marks[len(r.marks)-1]
	for _, side := range beaconSides {
		emit(goal.Sub(side), total)
		emit(goal.Add(side), total)
		emit(goal.Add(side).Add(beaconLift), total+beaconHeight)
		emit(goal.Sub(side).Add(beaconLift), total+beaconHeight)
	}

	quads := len(vertices) / routeFloatsPerQuad
	r.mesh = gpu.NewMesh(vertices, gpu.QuadIndices(quads),
		gpu.Attribute{Location: 0, Size: 3},
		gpu.Attribute{Location: 1, Size: 1})
}

func (r *Route) progress(next int, feet gocraft.Vec3d) float32 {
	if next <= 0 {
		return 0
	}
	if next >= len(r.points) {
		return r.marks[len(r.marks)-1]
	}

	position := mgl32.Vec3{
		float32(feet.X),
		float32(feet.Y) + routeLift,
		float32(feet.Z),
	}
	from := r.points[next-1]
	segment := r.points[next].Sub(from)

	length := segment.LenSqr()
	if length == 0 {
		return r.marks[next-1]
	}

	along := min(max(position.Sub(from).Dot(segment)/length, 0), 1)

	return r.marks[next-1] + along*segment.Len()
}

func (r *Route) clear() {
	if r.mesh != nil {
		r.mesh.Delete()
		r.mesh = nil
	}

	r.points = nil
	r.marks = nil
}
