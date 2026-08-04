package viewer

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	sensitivity = 0.15
	pitchLimit  = 89
)

// eye is where the viewer looks from and at. The bot moves once a tick and the
// window draws many times between, so the position is carried as the two ends of
// a glide rather than a single point.
type eye struct {
	camera Camera
	yaw    float32
	pitch  float32

	from     mgl32.Vec3
	to       mgl32.Vec3
	since    float64
	lastTick uint64
}

func newEye(spawn agent.Snapshot, aspect float32) eye {
	at := eyeOf(spawn.Position)

	return eye{
		camera: Camera{
			Up:     mgl32.Vec3{0, 1, 0},
			FOV:    fieldOfView,
			Aspect: aspect,
			Near:   nearPlane,
			Far:    farPlane,
		},
		yaw:      spawn.Yaw,
		pitch:    spawn.Pitch,
		from:     at,
		to:       at,
		lastTick: spawn.Tick,
	}
}

func (e *eye) aim(moved gpu.Point) {
	e.yaw += moved.X * sensitivity
	e.pitch = clamp(e.pitch+moved.Y*sensitivity, -pitchLimit, pitchLimit)
}

func (e *eye) facing() (yaw, pitch float32) {
	return e.yaw, e.pitch
}

// follow catches the camera up to the last tick and glides the rest of the way,
// so the world does not step at the tick rate
func (e *eye) follow(snapshot agent.Snapshot, now float64) {
	if snapshot.Tick != e.lastTick {
		e.lastTick = snapshot.Tick
		e.from = e.to
		e.to = eyeOf(snapshot.Position)
		e.since = now
	}

	alpha := float32(min((now-e.since)/codec.TickRate.Seconds(), 1))
	at := e.from.Add(e.to.Sub(e.from).Mul(alpha))

	e.camera.Position = at
	e.camera.Target = at.Add(direction(e.yaw, e.pitch))
}

// a minimised window reports no height at all, and dividing by it would leave the
// projection unusable for the rest of the session
func (e *eye) reframe(screen gpu.Rect) {
	if screen.Height() == 0 {
		return
	}

	e.camera.Aspect = screen.Width() / screen.Height()
}

func eyeOf(position gocraft.Vec3d) mgl32.Vec3 {
	return mgl32.Vec3{float32(position.X), float32(position.Y) + gocraft.EyeHeight, float32(position.Z)}
}

func direction(yaw, pitch float32) mgl32.Vec3 {
	y := float64(mgl32.DegToRad(yaw))
	p := float64(mgl32.DegToRad(pitch))

	return mgl32.Vec3{
		float32(-math.Sin(y) * math.Cos(p)),
		float32(-math.Sin(p)),
		float32(math.Cos(y) * math.Cos(p)),
	}
}

func clamp(value, low, high float32) float32 {
	return min(max(value, low), high)
}
