package viewer

import (
	gocraft "github.com/lrnxzz/go-craft"
)

const (
	// how long the camera has to rest on a block before it is marked
	dwellSeconds = 3.0

	// the camera never sits perfectly still, so resting is a small budget rather
	// than an exact match: a hand on the mouse drifts a pixel a frame
	dwellDrift = 0.35
	dwellTurn  = 1.5

	// once a block is marked the camera has to leave it before it can be marked
	// again, or resting on it would plant a point every three seconds forever
	dwellReach = 96.0
)

// dwell watches a loose camera for stillness. Holding the aim on one block long
// enough is how a point is set: no click, so both hands stay on flying.
type dwell struct {
	at      gocraft.Position
	holding bool
	since   float64

	from  gocraft.Vec3d
	yaw   float32
	pitch float32

	planted gocraft.Position
	spent   bool
}

// look reports the block being rested on and how far through the wait it is,
// from 0 to 1. It answers false when nothing is being held.
func (d *dwell) look(eye gocraft.Vec3d, yaw, pitch float32, hit gocraft.RayHit, sighted bool, now float64) (gocraft.Position, float64, bool) {
	moved := eye.Distance(d.from) > dwellDrift
	turned := abs(yaw-d.yaw) > dwellTurn || abs(pitch-d.pitch) > dwellTurn

	d.from, d.yaw, d.pitch = eye, yaw, pitch

	if !sighted {
		d.holding = false
		d.spent = false

		return gocraft.Position{}, 0, false
	}

	// leaving the block that was just planted arms it again
	if d.spent && hit.Block != d.planted {
		d.spent = false
	}

	if moved || turned || hit.Block != d.at || !d.holding {
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
	d.holding = false
	d.spent = false
}

func abs(value float32) float32 {
	if value < 0 {
		return -value
	}

	return value
}
