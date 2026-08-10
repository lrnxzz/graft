package viewer

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func aimedAt(at gocraft.Position) gocraft.RayHit {
	return gocraft.RayHit{
		Block: at,
	}
}

// Keeping the crosshair on a block is what sets a point.
func TestRestingLongEnoughSettles(t *testing.T) {
	var d dwell

	at := gocraft.At(10, 64, 10)

	d.look(aimedAt(at), true, 0)

	if _, settled := d.settled(dwellSeconds - 0.1); settled {
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

// The wait follows the block and not the camera. An earlier rule restarted it
// whenever the camera moved, and a single frame of flying moved it further than
// the budget allowed — so the wait could never finish while you were steering,
// which is the only time it is wanted.
func TestFlyingWhileAimedAtOneBlockStillSettles(t *testing.T) {
	var d dwell

	at := gocraft.At(10, 64, 10)

	d.look(aimedAt(at), true, 0)
	d.look(aimedAt(at), true, 0.5)
	d.look(aimedAt(at), true, 1.0)

	if _, settled := d.settled(dwellSeconds + 0.1); !settled {
		t.Fatal("flying while aimed at one block was mistaken for looking away")
	}
}

// crossing to another block starts the wait over, since that is the player
// choosing somewhere else
func TestCrossingToAnotherBlockRestartsTheWait(t *testing.T) {
	var d dwell

	d.look(aimedAt(gocraft.At(10, 64, 10)), true, 0)
	d.look(aimedAt(gocraft.At(11, 64, 10)), true, dwellSeconds-0.1)

	if _, settled := d.settled(dwellSeconds + 0.1); settled {
		t.Fatal("the wait carried over to a block it never rested on")
	}
}

// a settled block must not settle again while the crosshair stays on it, or one
// rest would plant a point forever
func TestASettledBlockDoesNotRepeat(t *testing.T) {
	var d dwell

	hit := aimedAt(gocraft.At(10, 64, 10))

	d.look(hit, true, 0)
	d.settled(dwellSeconds + 0.1)

	d.look(hit, true, dwellSeconds+0.2)

	if _, again := d.settled(2 * dwellSeconds); again {
		t.Fatal("the same block settled twice without the crosshair leaving it")
	}
}

// looking away and back arms it again
func TestLeavingTheBlockArmsItAgain(t *testing.T) {
	var d dwell

	here := aimedAt(gocraft.At(10, 64, 10))
	there := aimedAt(gocraft.At(20, 64, 20))

	d.look(here, true, 0)
	d.settled(dwellSeconds + 0.1)

	d.look(there, true, 4)
	d.look(here, true, 5)

	if _, again := d.settled(5 + dwellSeconds + 0.1); !again {
		t.Fatal("the block never armed again after the crosshair left it")
	}
}

// a click plants straight away, and arming has to leave the wait ready for the
// next block rather than mid-count on this one
func TestArmingClearsTheWait(t *testing.T) {
	var d dwell

	hit := aimedAt(gocraft.At(10, 64, 10))

	d.look(hit, true, 0)
	d.arm()

	if _, settled := d.settled(dwellSeconds + 0.1); settled {
		t.Fatal("a wait that was armed away still settled")
	}
}

// aiming at nothing holds nothing
func TestAimingAtSkyHoldsNothing(t *testing.T) {
	var d dwell

	if _, _, holding := d.look(gocraft.RayHit{}, false, 0); holding {
		t.Fatal("the sky was held")
	}

	if _, settled := d.settled(dwellSeconds + 1); settled {
		t.Fatal("the sky settled")
	}
}
