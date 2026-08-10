package viewer

import (
	gocraft "github.com/lrnxzz/go-craft"
)

const (
	// how long the crosshair has to stay on a block before it is marked
	dwellSeconds = 1.5

	// once a block is marked the crosshair has to leave it before it can be
	// marked again, or resting on it would plant a point forever
	dwellReach = 96.0
)

// dwell watches what the crosshair is on. Keeping it on one block long enough is
// how a point is set without reaching for a key.
//
// It follows the block and not the camera: an earlier version asked the camera
// to hold still, and a single frame of flying moved it further than the budget
// allowed, so the wait could never finish while you were steering.
type dwell struct {
	at      gocraft.Position
	holding bool
	since   float64

	planted gocraft.Position
	spent   bool
}

// look reports the block being rested on and how far through the wait it is,
// from 0 to 1. It answers false when nothing is being held.
func (d *dwell) look(hit gocraft.RayHit, sighted bool, now float64) (gocraft.Position, float64, bool) {
	if !sighted {
		d.holding = false
		d.spent = false

		return gocraft.Position{}, 0, false
	}

	// leaving the block that was just planted arms it again
	if d.spent && hit.Block != d.planted {
		d.spent = false
	}

	if hit.Block != d.at || !d.holding {
		d.at = hit.Block
		d.since = now
		d.holding = true

		return d.at, 0, !d.spent
	}

	held := (now - d.since) / dwellSeconds

	return d.at, min(held, 1), !d.spent
}

// settled reports the block the wait just completed on, once
func (d *dwell) settled(now float64) (gocraft.Position, bool) {
	if !d.holding || d.spent || now-d.since < dwellSeconds {
		return gocraft.Position{}, false
	}

	d.planted = d.at
	d.spent = true

	return d.at, true
}

func (d *dwell) forget() {
	d.arm()
}

// arm forgets the wait and the block it planted, so the next look starts over
func (d *dwell) arm() {
	d.holding = false
	d.spent = false
}
