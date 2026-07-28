package viewer

import (
	"context"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	sensitivity     = 0.15
	doubleTapWindow = 0.3
	swingSeconds    = 0.3
	pitchLimit      = 89
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

	if v.edges.key(gpu.KeyE).started {
		v.toggleInventory()
	}

	if v.inventoryOpen {
		v.bot.SetControls(gocraft.Controls{})
		v.pickSlot()

		return
	}

	if v.edges.key(gpu.KeyP).started {
		v.togglePathfinder()
	}

	if v.edges.key(gpu.KeyT).started {
		v.openChat("")

		return
	}
	if v.edges.key(gpu.KeySlash).started {
		v.openChat("/")

		return
	}

	now := v.window.Time()
	v.walk(now)
	v.aim()
	v.strike(ctx, now)
	v.hotkeys()
}

func (v *Viewer) walk(now float64) {
	forward := v.edges.key(gpu.KeyW)
	if forward.started {
		if now-v.lastW < doubleTapWindow {
			v.sprinting = true
		}

		v.lastW = now
	}
	if !forward.down {
		v.sprinting = false
	}

	v.bot.SetControls(gocraft.Controls{
		Forward: forward.down,
		Back:    v.window.Pressed(gpu.KeyS),
		Left:    v.window.Pressed(gpu.KeyA),
		Right:   v.window.Pressed(gpu.KeyD),
		Jump:    v.window.Pressed(gpu.KeySpace),
		Sprint:  v.sprinting || v.window.Pressed(gpu.KeyCtrl),
	})
}

func (v *Viewer) aim() {
	moved := v.window.CursorDelta()

	v.yaw += moved.X * sensitivity
	v.pitch = clamp(v.pitch+moved.Y*sensitivity, -pitchLimit, pitchLimit)
	v.bot.Look(v.yaw, v.pitch)
}

func (v *Viewer) strike(ctx context.Context, now float64) {
	digging := v.edges.button(gpu.ButtonLeft)
	switch {
	case digging.down:
		if digging.started || now-v.swungAt >= swingSeconds {
			v.swungAt = now
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
		v.swungAt = now
		_ = v.bot.Place(gocraft.BlockReach)
	}
}

func (v *Viewer) swing(now float64) float64 {
	since := now - v.swungAt
	if since >= swingSeconds {
		return 0
	}

	return since / swingSeconds
}

func (v *Viewer) togglePathfinder() {
	v.manual = !v.manual
	if v.manual {
		v.bot.Stop()
		v.chat.Push("pathfinder paused — press P to resume")

		return
	}

	v.chat.Push("pathfinder resumed")
}

func (v *Viewer) toggleInventory() {
	v.inventoryOpen = !v.inventoryOpen

	if v.inventoryOpen {
		v.window.ReleaseCursor()

		return
	}

	v.window.GrabCursor()
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
	if sendable {
		_ = v.bot.Chat(message)
	}
}

func (v *Viewer) pickSlot() {
	if !v.edges.button(gpu.ButtonLeft).started {
		return
	}

	slot, hit := v.hud.SlotAt(v.window.Cursor(), v.window.Viewport())
	if hit {
		_ = v.bot.ClickSlot(slot)
	}
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

func (v *Viewer) follow() {
	snapshot := v.bot.Snapshot()
	if snapshot.Tick != v.lastTick {
		v.lastTick = snapshot.Tick
		v.from = v.to
		v.to = eyeOf(snapshot.Position)
		v.since = v.window.Time()
	}

	alpha := float32(min((v.window.Time()-v.since)/gocraft.TickRate.Seconds(), 1))
	eye := v.from.Add(v.to.Sub(v.from).Mul(alpha))

	v.camera.Position = eye
	v.camera.Target = eye.Add(direction(v.yaw, v.pitch))
}

func eyeOf(position gocraft.Vec3d) mgl32.Vec3 {
	return mgl32.Vec3{float32(position.X), float32(position.Y) + gocraft.EyeHeight, float32(position.Z)}
}

func direction(yaw, pitch float32) mgl32.Vec3 {
	y := float64(mgl32.DegToRad(yaw))
	p := float64(mgl32.DegToRad(pitch))

	return mgl32.Vec3{
		float32(-math.Sin(y) * math.Cos(p)),
		float32(-math.Sin(p)),
		float32(math.Cos(y) * math.Cos(p)),
	}
}

func clamp(value, low, high float32) float32 {
	return min(max(value, low), high)
}
