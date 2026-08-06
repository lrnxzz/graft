package viewer

import (
	"context"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
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
		v.bot.SetControls(gocraft.Controls{})
		v.compose()

		return
	}

	if v.Showing() {
		v.bot.SetControls(gocraft.Controls{})
		v.drive()

		return
	}

	v.fire()

	// a bind may have just opened the chat or a menu, and neither wants the
	// keystroke that opened it to also reach the game
	if v.chat.Typing() || v.Showing() {
		v.bot.SetControls(gocraft.Controls{})

		return
	}

	now := v.window.Time()

	// out of the body the keys move the camera, and the bot keeps standing still
	if v.eye.away() {
		v.bot.SetControls(gocraft.Controls{})
		v.fly()
		v.aim()

		return
	}

	v.walk(now)
	v.aim()
	v.strike(ctx, now)
	v.hotkeys()
}

const (
	flySpeed  = 0.45
	flyFaster = 3
)

func (v *Viewer) fly() {
	held := func(key gpu.Key) float32 {
		if v.window.Pressed(key) {
			return 1
		}

		return 0
	}

	speed := float32(flySpeed)
	if v.window.Pressed(gpu.KeyCtrl) {
		speed *= flyFaster
	}

	v.eye.fly(
		held(gpu.KeyW)-held(gpu.KeyS),
		held(gpu.KeyD)-held(gpu.KeyA),
		held(gpu.KeySpace)-held(gpu.KeyShift),
		speed)
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

	held := gocraft.Controls{
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
				_, _ = v.bot.Dig(ctx, gocraft.BlockReach)
			}()
		}
	case digging.released:
		_ = v.bot.StopDigging()
	}

	if v.edges.button(gpu.ButtonRight).started {
		v.driving.swung(now)
		_ = v.bot.Place(gocraft.BlockReach)
	}
}

func (v *Viewer) toggleFreecam() {
	if v.Detached() {
		v.Attach()
		v.chat.Push("back in the body")

		return
	}

	v.Detach()
	v.chat.Push("out of the body — fly with WASD, hold still 3s on a block to mark it, F to return")
}

func (v *Viewer) togglePathfinder() {
	if v.driving.takeWheel() {
		v.bot.Stop()
		v.chat.Push("pathfinder paused — press P to resume")

		return
	}

	v.chat.Push("pathfinder resumed")
}

// the inventory is built fresh each time so its texture follows the Closer
// contract every other screen obeys
func (v *Viewer) openInventory() {
	screen, err := NewInventoryScreen(v.bot)
	if err != nil {
		v.chat.Push("inventory unavailable: " + err.Error())

		return
	}

	v.Open(screen)
}

func (v *Viewer) openChat(prefill string) {
	v.window.Typed()
	v.chat.Open(prefill)
}

func (v *Viewer) compose() {
	v.chat.Type(v.window.Typed())

	if v.edges.key(gpu.KeyBackspace).started {
		v.chat.Erase()
	}
	if v.edges.key(gpu.KeyEscape).started {
		v.chat.Cancel()
	}
	if !v.edges.key(gpu.KeyEnter).started {
		return
	}

	message, sendable := v.chat.Submit()
	if !sendable || v.stage.claimed(message) {
		return
	}

	_ = v.bot.Chat(message)
}

func (v *Viewer) hotkeys() {
	for index := range gocraft.HotbarSize {
		if v.edges.key(gpu.Digit(index)).started {
			_ = v.bot.SelectHotbar(index)
		}
	}

	scrolled := int(v.window.Scroll())
	if scrolled == 0 {
		return
	}

	held := v.bot.Inventory().HeldIndex() - scrolled
	_ = v.bot.SelectHotbar(wrap(held, gocraft.HotbarSize))
}

func wrap(index, size int) int {
	return ((index % size) + size) % size
}
