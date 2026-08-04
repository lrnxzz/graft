package viewer

const (
	doubleTapWindow = 0.3
	swingSeconds    = 0.3
)

// driving is what the player is doing at the controls and cannot be read from
// the bot: a sprint the double tap started, an arm still swinging, and whether
// the player took the wheel from the pathfinder.
type driving struct {
	sprinting bool
	tappedW   float64
	swungAt   float64
	manual    bool
}

// sprint reports whether to keep sprinting. Two taps of forward inside the
// window start it, and letting go of forward ends it.
func (d *driving) sprint(forward press, now float64) bool {
	if forward.started {
		if now-d.tappedW < doubleTapWindow {
			d.sprinting = true
		}

		d.tappedW = now
	}
	if !forward.down {
		d.sprinting = false
	}

	return d.sprinting
}

func (d *driving) swung(now float64) {
	d.swungAt = now
}

// swing is how far through the arm's arc the hand is, from 0 to 1, and 0 once it
// has come to rest
func (d *driving) swing(now float64) float64 {
	since := now - d.swungAt
	if since >= swingSeconds {
		return 0
	}

	return since / swingSeconds
}

func (d *driving) swinging(now float64) bool {
	return now-d.swungAt < swingSeconds
}

func (d *driving) takeWheel() bool {
	d.manual = !d.manual

	return d.manual
}
