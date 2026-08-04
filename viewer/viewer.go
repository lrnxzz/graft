package viewer

import (
	"context"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
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
	renderer *Renderer
	hand     *Hand
	chat     *Chat
	hud      *Hud
	edges    *edges
	camera   Camera
	bot      *agent.Agent
	yaw      float32
	pitch    float32

	overlayProgram *gpu.Program
	paintProgram   *gpu.Program
	canvas         Canvas
	painter        Painter

	world   []WorldLayer
	layers  []Layer
	screens []Screen
	binds   map[gpu.Key]func()

	// the atlases are shared by the renderer, the canvas and the hand, so the
	// viewer outlives all three and is what frees them
	tileset *Tileset
	iconset *Iconset
	font    *Font

	from     mgl32.Vec3
	to       mgl32.Vec3
	since    float64
	lastTick uint64

	sprinting bool
	lastW     float64
	swungAt   float64
	manual    bool
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
	tileset, err := LoadTileset()
	if err != nil {
		return nil, err
	}

	iconset, err := LoadIconset()
	if err != nil {
		return nil, err
	}

	font, err := LoadFont()
	if err != nil {
		return nil, err
	}

	renderer, err := NewRenderer(tileset)
	if err != nil {
		return nil, err
	}
	renderer.Build(bot.World())

	hand, err := NewHand(tileset, iconset)
	if err != nil {
		return nil, err
	}

	hud, err := NewHud(bot)
	if err != nil {
		return nil, err
	}

	overlay, err := gpu.NewProgram(hudVertexShader, hudFragmentShader)
	if err != nil {
		return nil, err
	}

	paint, err := gpu.NewProgram(paintVertexShader, paintFragmentShader)
	if err != nil {
		return nil, err
	}

	chat := NewChat(window.Time)

	echo := func(e agent.ChatReceived) {
		chat.Push(e.Line)
	}

	agent.On(bot, echo)

	spawn := bot.Snapshot()
	eye := eyeOf(spawn.Position)

	view := &Viewer{
		window:         window,
		renderer:       renderer,
		hand:           hand,
		chat:           chat,
		hud:            hud,
		overlayProgram: overlay,
		paintProgram:   paint,
		tileset:        tileset,
		iconset:        iconset,
		font:           font,
		edges:          newEdges(window),
		binds:          map[gpu.Key]func(){},
		bot:            bot,
		from:           eye,
		to:             eye,
		lastTick:       spawn.Tick,
		yaw:            spawn.Yaw,
		pitch:          spawn.Pitch,
		camera: Camera{
			Up:     mgl32.Vec3{0, 1, 0},
			FOV:    fieldOfView,
			Aspect: float32(defaultWidth) / float32(defaultHeight),
			Near:   nearPlane,
			Far:    farPlane,
		},
	}
	view.canvas.font = font
	view.canvas.icons = iconset

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
	v.world = append(v.world, layer)
}

// AddLayer registers an overlay drawn on top of the world, in screen space
func (v *Viewer) AddLayer(layer Layer) {
	v.layers = append(v.layers, layer)
}

// Open puts a menu on the stack: the bot stops moving, the cursor comes back and
// input goes to this screen until it is dismissed
func (v *Viewer) Open(screen Screen) {
	v.screens = append(v.screens, screen)
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
	if len(v.screens) == 0 {
		return
	}

	closeIfPossible(v.screens[len(v.screens)-1])
	v.screens = v.screens[:len(v.screens)-1]

	if len(v.screens) == 0 {
		v.window.GrabCursor()
	}
}

// Showing reports whether any menu is up, which is also when the game ignores
// movement and the crosshair is hidden
func (v *Viewer) Showing() bool {
	return len(v.screens) > 0
}

// Bind claims a key while no menu is up. It is how a plugin opens its own screen.
func (v *Viewer) Bind(key gpu.Key, action func()) {
	v.binds[key] = action
}

// Pick answers what the crosshair is pointing at, so a plugin can react to what
// the player is looking at without doing its own raycast
func (v *Viewer) Pick(reach float64) (gocraft.RayHit, bool) {
	return v.bot.Target(reach)
}

func (v *Viewer) Bot() *agent.Agent {
	return v.bot
}

// Chat is where a plugin reports to the player
func (v *Viewer) Chat() *Chat {
	return v.chat
}

func (v *Viewer) Manual() bool {
	return v.manual
}

func (v *Viewer) Close() {
	for _, layer := range v.world {
		closeIfPossible(layer)
	}
	for _, layer := range v.layers {
		closeIfPossible(layer)
	}
	for range v.screens {
		v.Dismiss()
	}

	v.hand.Close()
	v.renderer.Close()
	v.font.Close()
	v.overlayProgram.Delete()
	v.paintProgram.Delete()
	v.iconset.Delete()
	v.tileset.Delete()
	v.window.Close()
}

func (v *Viewer) frame() {
	now := v.window.Time()

	v.reframe()
	v.follow()
	v.hand.Update(v.bot.Inventory().Held())

	v.window.Clear(skyRed, skyGreen, skyBlue)
	v.renderer.Draw(v.camera)

	v.painter.reset(v.camera, now)
	for _, layer := range v.world {
		layer.DrawWorld(&v.painter)
	}
	v.painter.paint(v.paintProgram)

	v.hand.Draw(v.window, v.camera, now, v.swing(now))

	v.overlay(now)
}

// the projection has to follow the framebuffer, otherwise a resized window
// stretches the world; a minimised one reports no height at all
func (v *Viewer) reframe() {
	screen := v.window.Viewport()
	if screen.Height() == 0 {
		return
	}

	v.camera.Aspect = screen.Width() / screen.Height()
}

func (v *Viewer) overlay(now float64) {
	v.hud.aiming = !v.Showing()

	v.window.Overlay(true)
	defer v.window.Overlay(false)

	v.canvas.reset(v.window.Viewport(), v.window.Cursor(), now)
	for _, layer := range v.layers {
		layer.Draw(&v.canvas)
	}
	for _, screen := range v.screens {
		screen.Draw(&v.canvas)
	}
	v.canvas.paint(v.overlayProgram)
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
