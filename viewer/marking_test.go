package viewer

import (
	"testing"

	graft "github.com/lrnxzz/graft"
)

// The cage, the ring and the beam are the only clock the player gets, so the
// geometry has to answer to the wait rather than sit still while it fills.
func TestTheMarkingGrowsWithTheWait(t *testing.T) {
	var (
		marking Marking
		painter Painter
	)

	at := graft.At(10, 64, 10)

	marking.Aiming(at, 0.1, true)
	painter.reset(Camera{}, 0)
	marking.DrawWorld(&painter)
	early := len(painter.segments)

	marking.Aiming(at, 0.95, true)
	painter.reset(Camera{}, 0)
	marking.DrawWorld(&painter)
	late := len(painter.segments)

	if early == 0 {
		t.Fatal("nothing was drawn at the start of the wait")
	}
	if late <= early {
		t.Errorf("the marking drew %d at the start and %d near the end", early, late)
	}
}

// nothing is drawn when the camera rests on nothing, or the last block aimed at
// would keep glowing after the player looked away
func TestNothingIsDrawnWithoutAHold(t *testing.T) {
	var (
		marking Marking
		painter Painter
	)

	marking.Aiming(graft.At(10, 64, 10), 0.5, false)
	painter.reset(Camera{}, 0)
	marking.DrawWorld(&painter)

	if len(painter.segments) != 0 {
		t.Fatalf("%d segments were drawn for nothing", len(painter.segments))
	}
}

// the flash is one frame: it fires on the frame a point lands and is gone after
func TestTheFlashLastsOneFrame(t *testing.T) {
	var (
		marking Marking
		painter Painter
	)

	marking.Aiming(graft.At(10, 64, 10), 1, true)
	marking.Settled()

	if !marking.settling {
		t.Fatal("the flash never armed")
	}

	painter.reset(Camera{}, 0)
	marking.DrawWorld(&painter)

	if marking.settling {
		t.Fatal("the flash outlived the frame it was drawn on")
	}
}
