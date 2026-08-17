package viewer

import (
	"math"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/viewer/gpu"
)

const (
	// the cage starts this far out and closes onto the block as the wait fills,
	// so the pull of it reads before any number would
	cageReach = 2.6
	cageSpin  = 2.2

	// the beam stands tall enough to be found from across the world
	beamHeight = 40
	beamPulse  = 3.5

	ringSegments = 24
	ringRadius   = 0.9
	ringRise     = 1.6
)

// Marking is what the player sees while resting the aim on a block: a cage that
// closes, a ring that spins and a beam that stands. It draws only the block being
// settled on — the points already planted are the plugin's to draw.
type Marking struct {
	at       graft.Position
	held     float64
	showing  bool
	settling bool
}

// Aiming is called by the frame with what the camera rests on and how far through
// the wait it is
func (m *Marking) Aiming(at graft.Position, held float64, showing bool) {
	m.at, m.held, m.showing = at, held, showing
}

// Settled flashes the moment a point is planted
func (m *Marking) Settled() {
	m.settling = true
}

func (m *Marking) DrawWorld(painter *Painter) {
	if !m.showing {
		return
	}

	now := painter.Time()
	middle := m.at.Center()

	m.cage(painter, middle)
	m.ring(painter, middle, now)
	m.beam(painter, middle, now)

	m.settling = false
}

// the cage is eight struts reaching in from the corners of a shrinking cube: at
// the start they are far out and faint, at the end they touch the block
func (m *Marking) cage(painter *Painter, middle graft.Vec3d) {
	reach := cageReach * (1 - m.held)
	tint := gpu.RGBA(0.49, 0.83, 0.99, float32(0.25+0.75*m.held))

	if m.settling {
		tint = gpu.RGBA(0.6, 1, 0.7, 1)
	}

	for _, corner := range corners {
		far := middle.Offset(corner.X*(0.5+reach), corner.Y*(0.5+reach), corner.Z*(0.5+reach))
		near := middle.Offset(corner.X*0.62, corner.Y*0.62, corner.Z*0.62)

		painter.Line(far, near, tint)
	}

	painter.Box(graft.Box(middle.Offset(-0.52, -0.52, -0.52), middle.Offset(0.52, 0.52, 0.52)), tint)
}

// the ring turns while the wait fills and closes it: an arc that grows to a full
// circle is the clock, read without a number
func (m *Marking) ring(painter *Painter, middle graft.Vec3d, now float64) {
	filled := int(float64(ringSegments) * m.held)
	if filled < 1 {
		return
	}

	turn := now * cageSpin
	tint := gpu.RGBA(0.49, 0.83, 0.99, 0.95)

	for segment := range filled {
		from := ringAt(middle, turn, segment)
		to := ringAt(middle, turn, segment+1)

		painter.Line(from, to, tint)
	}
}

func ringAt(middle graft.Vec3d, turn float64, segment int) graft.Vec3d {
	angle := turn + float64(segment)*2*math.Pi/ringSegments

	return middle.Offset(math.Cos(angle)*ringRadius, ringRise, math.Sin(angle)*ringRadius)
}

// the beam is how a point is found again from far away, and it breathes so it is
// never mistaken for terrain
func (m *Marking) beam(painter *Painter, middle graft.Vec3d, now float64) {
	breath := 0.55 + 0.45*math.Sin(now*beamPulse)
	tint := gpu.RGBA(0.49, 0.83, 0.99, float32(breath*m.held))

	top := middle.Offset(0, beamHeight, 0)

	painter.Line(middle, top, tint)
	painter.Line(middle.Offset(0.2, 0, 0.2), top.Offset(0.2, 0, 0.2), tint)
	painter.Line(middle.Offset(-0.2, 0, -0.2), top.Offset(-0.2, 0, -0.2), tint)
}

var corners = [...]graft.Vec3d{
	graft.Vec3(1, 1, 1),
	graft.Vec3(1, 1, -1),
	graft.Vec3(1, -1, 1),
	graft.Vec3(1, -1, -1),
	graft.Vec3(-1, 1, 1),
	graft.Vec3(-1, 1, -1),
	graft.Vec3(-1, -1, 1),
	graft.Vec3(-1, -1, -1),
}
