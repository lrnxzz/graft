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
	flown(&e, 1, 0, 0, 0.5)

	if e.at.Len() == 0 {
		t.Fatal("flying forward moved nothing")
	}
}

// flown holds the keys for a stretch of seconds, since one call only starts the
// clock and the camera eases into its pace over the frames that follow
func flown(e *eye, forward, strafe, rise float32, seconds float64) {
	const frame = 1.0 / 60

	for at := 0.0; at <= seconds; at += frame {
		e.fly(forward, strafe, rise, 1, at)
	}
}

// rising leaves the horizon alone, or looking down would turn every ascent into
// a dive
func TestRisingKeepsTheHorizon(t *testing.T) {
	var e eye
	e.leave()
	e.aim(gpu.At(0, 40))

	before := e.at
	flown(&e, 0, 0, 1, 0.5)

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
	flown(&e, 1, 1, 1, 0.5)

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

// The camera flies in blocks a second, not blocks a frame. It used to move a
// fixed step per frame, so the same stick carried you twice as far on a fast
// machine — which is what made steering it a guess.
func TestTheSameStretchFliesTheSameDistance(t *testing.T) {
	fly := func(frame float64) float32 {
		var e eye
		e.leave()
		e.aim(gpu.At(90, 0))

		for at := 0.0; at <= 1.0; at += frame {
			e.fly(1, 0, 0, 1, at)
		}

		return e.at.Len()
	}

	slow, fast := fly(1.0/30), fly(1.0/144)

	if drift := abs32(slow-fast) / max(slow, fast); drift > 0.05 {
		t.Errorf("30fps flew %.2f and 144fps flew %.2f, %.0f%% apart", slow, fast, drift*100)
	}
}

// a stall must not fling the camera across the world when the frames come back
func TestAStallDoesNotFlingTheCamera(t *testing.T) {
	var e eye
	e.leave()
	e.aim(gpu.At(90, 0))

	e.fly(1, 0, 0, 1, 0)
	e.fly(1, 0, 0, 1, 30)

	if flung := e.at.Len(); flung > flyCruise {
		t.Errorf("a thirty second stall moved the camera %.0f blocks", flung)
	}
}

// pressing two keys asks for one speed, or a diagonal would be the fast way to
// cross anything
func TestADiagonalIsNotAShortcut(t *testing.T) {
	straight, corner := eye{}, eye{}
	straight.leave()
	corner.leave()

	flown(&straight, 1, 0, 0, 1)
	flown(&corner, 1, 1, 0, 1)

	if corner.at.Len() > straight.at.Len()*1.05 {
		t.Errorf("a diagonal flew %.2f against %.2f straight", corner.at.Len(), straight.at.Len())
	}
}

// the wheel moves the cruise itself, and it has to stop at both ends
func TestTheCruiseStopsAtBothEnds(t *testing.T) {
	var e eye
	e.leave()

	for range 200 {
		e.pace(1)
	}
	if e.cruising() > flyFastest {
		t.Errorf("the cruise ran past the top at %.1f", e.cruising())
	}

	for range 400 {
		e.pace(-1)
	}
	if e.cruising() < flySlowest {
		t.Errorf("the cruise ran past the bottom at %.1f", e.cruising())
	}
}

func abs32(value float32) float32 {
	if value < 0 {
		return -value
	}

	return value
}
