package viewer

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

// Resting the aim on a block is what sets a point, so the wait has to survive the
// drift of a hand on a mouse and still restart the moment the camera really moves.
func TestRestingLongEnoughSettles(t *testing.T) {
	var d dwell

	at := gocraft.At(10, 64, 10)
	eye := gocraft.Vec3(0, 70, 0)

	hit := gocraft.RayHit{
		Block: at,
	}

	d.look(eye, 0, 0, hit, true, 0)

	_, settled := d.settled(dwellSeconds - 0.1)
	if settled {
		t.Fatal("settled before the wait was over")
	}

	block, settled := d.settled(dwellSeconds + 0.1)
	if !settled {
		t.Fatal("the wait ran out and nothing settled")
	}
	if block != at {
		t.Errorf("settled on %v, want %v", block, at)
	}
}

// a settled block must not settle again while the camera stays on it, or one
// rest would plant a point every three seconds
func TestASettledBlockDoesNotRepeat(t *testing.T) {
	var d dwell

	hit := gocraft.RayHit{
		Block: gocraft.At(10, 64, 10),
	}
	eye := gocraft.Vec3(0, 70, 0)

	d.look(eye, 0, 0, hit, true, 0)
	d.settled(dwellSeconds + 0.1)

	d.look(eye, 0, 0, hit, true, dwellSeconds+0.2)

	if _, again := d.settled(2 * dwellSeconds); again {
		t.Fatal("the same block settled twice without the camera leaving it")
	}
}

// looking away and back arms it again
func TestLeavingTheBlockArmsItAgain(t *testing.T) {
	var d dwell

	here := gocraft.RayHit{
		Block: gocraft.At(10, 64, 10),
	}
	there := gocraft.RayHit{
		Block: gocraft.At(20, 64, 20),
	}
	eye := gocraft.Vec3(0, 70, 0)

	d.look(eye, 0, 0, here, true, 0)
	d.settled(dwellSeconds + 0.1)

	d.look(eye, 0, 0, there, true, 4)
	d.look(eye, 0, 0, here, true, 5)

	if _, again := d.settled(5 + dwellSeconds + 0.1); !again {
		t.Fatal("the block never armed again after the camera left it")
	}
}

// moving the camera restarts the wait, or a point would be planted mid-flight
func TestMovingRestartsTheWait(t *testing.T) {
	var d dwell

	hit := gocraft.RayHit{
		Block: gocraft.At(10, 64, 10),
	}

	d.look(gocraft.Vec3(0, 70, 0), 0, 0, hit, true, 0)
	d.look(gocraft.Vec3(0, 70, 5), 0, 0, hit, true, 2)

	if _, settled := d.settled(dwellSeconds + 0.1); settled {
		t.Fatal("the wait survived the camera flying away")
	}
}

// turning restarts it too, even standing in one spot
func TestTurningRestartsTheWait(t *testing.T) {
	var d dwell

	hit := gocraft.RayHit{
		Block: gocraft.At(10, 64, 10),
	}
	eye := gocraft.Vec3(0, 70, 0)

	d.look(eye, 0, 0, hit, true, 0)
	d.look(eye, 40, 0, hit, true, 2)

	if _, settled := d.settled(dwellSeconds + 0.1); settled {
		t.Fatal("the wait survived the camera turning")
	}
}

// a hand resting on a mouse drifts; that must not count as flying away
func TestASteadyHandStillCounts(t *testing.T) {
	var d dwell

	hit := gocraft.RayHit{
		Block: gocraft.At(10, 64, 10),
	}

	d.look(gocraft.Vec3(0, 70, 0), 0, 0, hit, true, 0)
	d.look(gocraft.Vec3(0.05, 70, 0.05), 0.2, -0.1, hit, true, 1)
	d.look(gocraft.Vec3(0.08, 70, 0.02), 0.4, -0.2, hit, true, 2)

	if _, settled := d.settled(dwellSeconds + 0.1); !settled {
		t.Fatal("a steady hand was mistaken for flying")
	}
}

// aiming at nothing holds nothing
func TestAimingAtSkyHoldsNothing(t *testing.T) {
	var d dwell

	_, _, holding := d.look(gocraft.Vec3(0, 70, 0), 0, 0, gocraft.RayHit{}, false, 0)
	if holding {
		t.Fatal("the sky was held")
	}

	if _, settled := d.settled(dwellSeconds + 1); settled {
		t.Fatal("the sky settled")
	}
}
