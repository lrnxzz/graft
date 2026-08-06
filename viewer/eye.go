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

	// loose is the camera out of the body: it stops following the bot and takes
	// movement of its own. Nothing of it reaches the server, which never learns
	// the player looked away.
	loose bool
	at    mgl32.Vec3
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

// leave lifts the camera off the body, starting where the body was looking
func (e *eye) leave() {
	e.at = e.camera.Position
	e.loose = true
}

func (e *eye) enter() {
	e.loose = false
}

func (e *eye) away() bool {
	return e.loose
}

func (e *eye) eye() gocraft.Vec3d {
	at := e.camera.Position

	return gocraft.Vec3(float64(at.X()), float64(at.Y()), float64(at.Z()))
}

func (e *eye) forward() gocraft.Vec3d {
	facing := direction(e.yaw, e.pitch)

	return gocraft.Vec3(float64(facing.X()), float64(facing.Y()), float64(facing.Z()))
}

// fly moves a loose camera. Forward is where it looks, so flying and aiming are
// the same gesture, and rise leaves the horizon alone.
func (e *eye) fly(forward, strafe, rise, speed float32) {
	if !e.loose {
		return
	}

	facing := direction(e.yaw, e.pitch)
	side := mgl32.Vec3{-facing.Z(), 0, facing.X()}.Normalize()

	e.at = e.at.
		Add(facing.Mul(forward * speed)).
		Add(side.Mul(strafe * speed)).
		Add(mgl32.Vec3{0, rise * speed, 0})
}

// follow catches the camera up to the last tick and glides the rest of the way,
// so the world does not step at the tick rate
func (e *eye) follow(snapshot agent.Snapshot, now float64) {
	if e.loose {
		e.camera.Position = e.at
		e.camera.Target = e.at.Add(direction(e.yaw, e.pitch))

		return
	}

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
