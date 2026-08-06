package viewer

import (
	"context"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
	remeshEvery   = 15

	fieldOfView = 70
	nearPlane   = 0.1
	farPlane    = 2000

	// vanilla overworld sky at noon
	skyRed   = 0.53
	skyGreen = 0.71
	skyBlue  = 0.92
)

// Viewer is the host a plugin extends: it owns the window, the world render and
// the frame, and hands out drawing time through layers, menus through screens and
// keys through binds. Everything it draws itself goes through those same seams.
type Viewer struct {
	window   *gpu.Window
	bot      *agent.Agent
	renderer *Renderer
	hand     *Hand
	chat     *Chat
	hud      *Hud
	edges    *edges

	eye      eye
	assets   *assets
	surfaces surfaces
	stage    stage
	driving  driving

	dwell   dwell
	marking Marking
	settled []func(gocraft.Position)
}

func New(bot *agent.Agent, visible bool) (*Viewer, error) {
	window, err := gpu.OpenWindow("gocraft", defaultWidth, defaultHeight, visible)
	if err != nil {
		return nil, err
	}

	view, err := attach(window, bot)
	if err != nil {
		window.Close()

		return nil, err
	}

	return view, nil
}

func attach(window *gpu.Window, bot *agent.Agent) (*Viewer, error) {
	loaded, err := loadAssets()
	if err != nil {
		return nil, err
	}

	drawn, err := newSurfaces(loaded)
	if err != nil {
		return nil, err
	}

	renderer, err := NewRenderer(loaded.tileset)
	if err != nil {
		return nil, err
	}
	renderer.Build(bot.World())

	hand, err := NewHand(loaded.tileset, loaded.iconset)
	if err != nil {
		return nil, err
	}

	hud, err := NewHud(bot)
	if err != nil {
		return nil, err
	}

	chat := NewChat(window.Time)

	echo := func(e agent.ChatReceived) {
		chat.Push(e.Line)
	}

	agent.On(bot, echo)

	view := &Viewer{
		window:   window,
		bot:      bot,
		renderer: renderer,
		hand:     hand,
		chat:     chat,
		hud:      hud,
		edges:    newEdges(window),
		eye:      newEye(bot.Snapshot(), float32(defaultWidth)/float32(defaultHeight)),
		assets:   loaded,
		surfaces: drawn,
		stage:    newStage(),
	}

	if err := view.installDefaults(); err != nil {
		return nil, err
	}

	return view, nil
}

// the viewer's own route, crack and hud are registered through the public API,
// so a plugin's layer is never a second-class citizen
func (v *Viewer) installDefaults() error {
	crack, err := NewCrack(v.bot)
	if err != nil {
		return err
	}

	route, err := NewRoute(v.bot)
	if err != nil {
		return err
	}

	v.AddWorldLayer(crack)
	v.AddWorldLayer(route)
	v.AddWorldLayer(&v.marking)
	v.AddLayer(v.hud)
	v.AddLayer(v.chat)

	v.Bind(gpu.KeyE, v.openInventory)
	v.Bind(gpu.KeyP, v.togglePathfinder)
	say := func() {
		v.openChat("")
	}
	command := func() {
		v.openChat("/")
	}

	v.Bind(gpu.KeyT, say)
	v.Bind(gpu.KeySlash, command)

	return nil
}

// AddWorldLayer registers geometry drawn inside the world, with depth still on
func (v *Viewer) AddWorldLayer(layer WorldLayer) {
	v.stage.addWorldLayer(layer)
}

// AddLayer registers an overlay drawn on top of the world, in screen space
func (v *Viewer) AddLayer(layer Layer) {
	v.stage.addLayer(layer)
}

// Open puts a menu on the stack: the bot stops moving, the cursor comes back and
// input goes to this screen until it is dismissed
func (v *Viewer) Open(screen Screen) {
	v.stage.open(screen)
	v.window.ReleaseCursor()
}

// Viewport is the drawable area, which a screen needs before its first frame if
// it has to size something to the window up front
func (v *Viewer) Viewport() gpu.Rect {
	return v.window.Viewport()
}

// Dismiss closes the topmost menu, handing control back to the game when the
// stack empties
func (v *Viewer) Dismiss() {
	if v.stage.dismiss() {
		v.window.GrabCursor()
	}
}

// Showing reports whether any menu is up, which is also when the game ignores
// movement and the crosshair is hidden
func (v *Viewer) Showing() bool {
	return v.stage.showing()
}

// Bind claims a key while no menu is up. It is how a plugin opens its own screen.
func (v *Viewer) Bind(key gpu.Key, action func()) {
	v.stage.bind(key, action)
}

// Intercept offers every line the player submits to a handler before the server
// sees it. Answering true keeps the line here, which is how a local command
// shadows a server one without hiding the ones nobody claims.
func (v *Viewer) Intercept(handle func(line string) bool) {
	v.stage.intercept(handle)
}

// Pick answers what the crosshair is pointing at. Out of the body that is the
// camera's own ray, not the bot's, so a plugin marking a spot marks the one the
// player can see rather than the one the body happens to face.
func (v *Viewer) Pick(reach float64) (gocraft.RayHit, bool) {
	if !v.eye.away() {
		return v.bot.Target(reach)
	}

	return v.bot.World().Raycast(v.eye.eye(), v.eye.forward(), reach, blocks.Solid)
}

// Detach lifts the camera off the body and Attach puts it back. The bot stops
// taking movement while it is away, and the server is never told.
func (v *Viewer) Detach() {
	v.eye.leave()
	v.dwell.forget()
	v.bot.SetControls(gocraft.Controls{})
}

func (v *Viewer) Attach() {
	v.eye.enter()
	v.dwell.forget()
	v.marking.Aiming(gocraft.Position{}, 0, false)
}

// Settles registers what happens when the player rests the aim on a block long
// enough for it to be marked. It is how a point is set with both hands still on
// flying, and it only ever fires while the camera is out of the body.
func (v *Viewer) Settles(handle func(at gocraft.Position)) {
	v.settled = append(v.settled, handle)
}

// rest feeds the wait with what the camera is looking at, and reports the block
// the moment it completes
func (v *Viewer) rest(now float64) {
	if !v.eye.away() {
		return
	}

	hit, sighted := v.Pick(dwellReach)
	yaw, pitch := v.eye.facing()

	at, held, showing := v.dwell.look(v.eye.eye(), yaw, pitch, hit, sighted, now)
	v.marking.Aiming(at, held, showing)

	block, done := v.dwell.settled(now)
	if !done {
		return
	}

	v.marking.Settled()

	for _, handle := range v.settled {
		handle(block)
	}
}

func (v *Viewer) Detached() bool {
	return v.eye.away()
}

func (v *Viewer) Bot() *agent.Agent {
	return v.bot
}

// Chat is where a plugin reports to the player
func (v *Viewer) Chat() *Chat {
	return v.chat
}

func (v *Viewer) Manual() bool {
	return v.driving.manual
}

func (v *Viewer) Close() {
	v.stage.close()
	v.surfaces.close()
	v.hand.Close()
	v.renderer.Close()
	v.assets.close()
	v.window.Close()
}

func (v *Viewer) frame() {
	now := v.window.Time()

	v.eye.reframe(v.window.Viewport())
	v.eye.follow(v.bot.Snapshot(), now)
	v.rest(now)
	v.hand.Update(v.bot.Inventory().Held())
	v.hud.Aiming(!v.Showing())

	v.window.Clear(skyRed, skyGreen, skyBlue)
	v.renderer.Draw(v.eye.camera)

	v.surfaces.world(v.eye.camera, now, v.stage.drawWorld)
	v.hand.Draw(v.window, v.eye.camera, now, v.driving.swing(now))
	v.surfaces.over(v.window, now, v.stage.draw)
}

func (v *Viewer) Run(ctx context.Context) {
	defer v.Close()

	v.window.GrabCursor()
	for frame := 0; !v.window.ShouldClose() && ctx.Err() == nil; frame++ {
		if frame%remeshEvery == 0 {
			v.renderer.Build(v.bot.World())
		}

		v.renderer.Collect()
		v.control(ctx)
		v.frame()
		v.window.Present()
	}
}

func (v *Viewer) Screenshot(path string) error {
	defer v.Close()

	v.renderer.Flush()
	v.frame()

	return v.window.Capture(path)
}
