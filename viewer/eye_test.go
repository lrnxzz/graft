package viewer

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

// Out of the body the camera stops following the bot. Without this the freecam
// would be dragged back to the body on every tick, which is once every 50ms.
func TestALooseCameraIgnoresTheBot(t *testing.T) {
	var e eye
	e.camera.Position = mgl32.Vec3{10, 64, 10}
	e.leave()

	walked := agent.Snapshot{
		Tick:     7,
		Position: gocraft.Vec3(200, 64, 200),
	}

	e.follow(walked, 1)

	if e.camera.Position != (mgl32.Vec3{10, 64, 10}) {
		t.Fatalf("a loose camera moved to %v when the bot did", e.camera.Position)
	}

	e.enter()
	e.follow(walked, 1)

	if e.camera.Position == (mgl32.Vec3{10, 64, 10}) {
		t.Fatal("back in the body the camera did not follow")
	}
}

// forward is where the camera looks, so flying and aiming are one gesture
func TestFlyingFollowsTheAim(t *testing.T) {
	var e eye
	e.leave()

	e.aim(gpu.At(90, 0))
	e.fly(1, 0, 0, 1)

	if e.at.Len() == 0 {
		t.Fatal("flying forward moved nothing")
	}
}

// rising leaves the horizon alone, or looking down would turn every ascent into
// a dive
func TestRisingKeepsTheHorizon(t *testing.T) {
	var e eye
	e.leave()
	e.aim(gpu.At(0, 40))

	before := e.at
	e.fly(0, 0, 1, 1)

	if e.at.X() != before.X() || e.at.Z() != before.Z() {
		t.Errorf("rising moved sideways: %v then %v", before, e.at)
	}
	if e.at.Y() <= before.Y() {
		t.Error("rising did not go up")
	}
}

// a camera still in the body must not drift when the keys are read
func TestFlyingDoesNothingInTheBody(t *testing.T) {
	var e eye
	e.fly(1, 1, 1, 10)

	if e.at != (mgl32.Vec3{}) {
		t.Fatalf("the body flew to %v", e.at)
	}
}

// the aim is clamped so the camera can never roll past straight up or down
func TestTheAimStopsAtTheZenith(t *testing.T) {
	var e eye
	e.aim(gpu.At(0, 10000))

	if e.pitch > pitchLimit {
		t.Errorf("pitch = %v, want no more than %v", e.pitch, pitchLimit)
	}

	e.aim(gpu.At(0, -20000))

	if e.pitch < -pitchLimit {
		t.Errorf("pitch = %v, want no less than %v", e.pitch, -pitchLimit)
	}
}
