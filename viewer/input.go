package viewer

import (
	"context"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/viewer/gpu"
)

type press struct {
	down     bool
	started  bool
	released bool
}

type edges struct {
	window  *gpu.Window
	keys    map[gpu.Key]bool
	buttons map[gpu.Button]bool
}

func newEdges(window *gpu.Window) *edges {
	return &edges{
		window:  window,
		keys:    map[gpu.Key]bool{},
		buttons: map[gpu.Button]bool{},
	}
}

func sample(down, held bool) press {
	return press{
		down:     down,
		started:  down && !held,
		released: !down && held,
	}
}

func (e *edges) key(key gpu.Key) press {
	down := e.window.Pressed(key)
	sampled := sample(down, e.keys[key])
	e.keys[key] = down

	return sampled
}

func (e *edges) button(button gpu.Button) press {
	down := e.window.Clicking(button)
	sampled := sample(down, e.buttons[button])
	e.buttons[button] = down

	return sampled
}

func (v *Viewer) control(ctx context.Context) {
	if v.chat.Typing() {
		v.bot.SetControls(graft.Controls{})
		v.compose()

		return
	}

	if v.Showing() {
		v.bot.SetControls(graft.Controls{})
		v.drive()

		return
	}

	v.fire()

	// a bind may have just opened the chat or a menu, and neither wants the
	// keystroke that opened it to also reach the game
	if v.chat.Typing() || v.Showing() {
		v.bot.SetControls(graft.Controls{})

		return
	}

	now := v.window.Time()

	// out of the body the keys move the camera, and the bot keeps standing still
	if v.eye.away() {
		v.bot.SetControls(graft.Controls{})
		v.fly(now)
		v.aim()
		v.plant()

		return
	}

	v.walk(now)
	v.aim()
	v.strike(ctx, now)
	v.hotkeys()
}

// Ctrl covers ground and Alt lines up on a block. The wheel moves the cruise
// itself, so the pace you settle on survives letting go of the keys.
const (
	flyBoost = 3.5
	flyCreep = 0.22
)

func (v *Viewer) fly(now float64) {
	held := func(key gpu.Key) float32 {
		if v.window.Pressed(key) {
			return 1
		}

		return 0
	}

	if scrolled := v.window.Scroll(); scrolled != 0 {
		v.eye.pace(scrolled)
	}

	pace := float32(1)
	switch {
	case v.window.Pressed(gpu.KeyCtrl):
		pace = flyBoost
	case v.window.Pressed(gpu.KeyAlt):
		pace = flyCreep
	}

	v.eye.fly(
		held(gpu.KeyW)-held(gpu.KeyS),
		held(gpu.KeyD)-held(gpu.KeyA),
		held(gpu.KeySpace)-held(gpu.KeyShift),
		pace,
		now)
}

// plant marks the block under the crosshair on a click. Waiting works too, but a
// click is what a hand already on the mouse reaches for first.
func (v *Viewer) plant() {
	if !v.edges.button(gpu.ButtonLeft).started {
		return
	}

	hit, sighted := v.Pick(dwellReach)
	if !sighted {
		return
	}

	v.Plant(hit.Block)
}

// fire runs whatever bind claimed a key this frame, the viewer's own included
func (v *Viewer) fire() {
	v.stage.fire(v.edges)
}

// drive hands the frame to the menu on top; escape and the key that opened it
// always dismiss, so a screen can never trap the player
func (v *Viewer) drive() {
	if v.edges.key(gpu.KeyEscape).started {
		v.Dismiss()

		return
	}

	top := v.stage.top()

	// the wheel accumulates whether or not anyone reads it, so it is drained even
	// when the menu ignores it, or the hotbar jumps the moment the menu closes
	scrolled := v.window.Scroll()

	wheeled, turns := top.(Scroller)
	if turns && scrolled != 0 {
		wheeled.Scroll(scrolled)
	}

	if v.edges.button(gpu.ButtonLeft).started {
		top.Click(v.window.Cursor(), v.window.Viewport())
	}

	for _, key := range gpu.Keys {
		if key == gpu.KeyEscape {
			continue
		}
		if v.edges.key(key).started {
			top.Key(key)
		}
	}
}

func (v *Viewer) walk(now float64) {
	forward := v.edges.key(gpu.KeyW)
	sprinting := v.driving.sprint(forward, now)

	held := graft.Controls{
		Forward: forward.down,
		Back:    v.window.Pressed(gpu.KeyS),
		Left:    v.window.Pressed(gpu.KeyA),
		Right:   v.window.Pressed(gpu.KeyD),
		Jump:    v.window.Pressed(gpu.KeySpace),
		Sprint:  sprinting || v.window.Pressed(gpu.KeyCtrl),
	}

	v.bot.SetControls(held)
}

func (v *Viewer) aim() {
	v.eye.aim(v.window.CursorDelta())

	// a loose camera turns nobody's head: the body keeps the aim it had
	if v.eye.away() {
		return
	}

	v.bot.Look(v.eye.facing())
}

func (v *Viewer) strike(ctx context.Context, now float64) {
	digging := v.edges.button(gpu.ButtonLeft)
	switch {
	case digging.down:
		if digging.started || !v.driving.swinging(now) {
			v.driving.swung(now)
		}
		if !v.bot.Digging() {
			// the frame must not wait on the block, and the crack overlay
			// already shows the progress, so the outcome goes nowhere
			go func() {
				_, _ = v.bot.Dig(ctx, graft.BlockReach)
			}()
		}
	case digging.released:
		_ = v.bot.StopDigging()
	}

	if v.edges.button(gpu.ButtonRight).started {
		v.driving.swung(now)
		_ = v.bot.Place(graft.BlockReach)
	}
}

func (v *Viewer) compose() {
	v.chat.Type(v.window.Typed())

	if v.edges.key(gpu.KeyBackspace).started {
		v.chat.Erase()
	}
	if v.edges.key(gpu.KeyEscape).started {
		v.chat.Cancel()
		v.suggesting.Clear()

		return
	}

	if v.suggest() {
		return
	}

	if !v.edges.key(gpu.KeyEnter).started {
		return
	}

	v.suggesting.Clear()

	message, sendable := v.chat.Submit()
	if !sendable || v.stage.claimed(message) {
		return
	}

	_ = v.bot.Chat(message)
}

func (v *Viewer) hotkeys() {
	for index := range graft.HotbarSize {
		if v.edges.key(gpu.Digit(index)).started {
			_ = v.bot.SelectHotbar(index)
		}
	}

	scrolled := int(v.window.Scroll())
	if scrolled == 0 {
		return
	}

	held := v.bot.Inventory().HeldIndex() - scrolled
	_ = v.bot.SelectHotbar(wrap(held, graft.HotbarSize))
}

func wrap(index, size int) int {
	return ((index % size) + size) % size
}
