package viewer

import (
	"context"

	"github.com/go-gl/mathgl/mgl32"
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

type Viewer struct {
	window   *gpu.Window
	renderer *Renderer
	route    *Route
	hud      *Hud
	hand     *Hand
	chat     *Chat
	crack    *Crack
	edges    *edges
	camera   Camera
	bot      *agent.Agent
	yaw      float32
	pitch    float32

	from     mgl32.Vec3
	to       mgl32.Vec3
	since    float64
	lastTick uint64

	sprinting     bool
	lastW         float64
	swungAt       float64
	inventoryOpen bool
	manual        bool
}

func (v *Viewer) Manual() bool {
	return v.manual
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

	renderer, err := NewRenderer(tileset)
	if err != nil {
		return nil, err
	}
	renderer.Build(bot.World())

	route, err := NewRoute()
	if err != nil {
		return nil, err
	}

	iconset, err := LoadIconset()
	if err != nil {
		return nil, err
	}

	hud, err := NewHud(iconset)
	if err != nil {
		return nil, err
	}

	hand, err := NewHand(tileset, iconset)
	if err != nil {
		return nil, err
	}

	chat, err := NewChat(window.Time)
	if err != nil {
		return nil, err
	}
	bot.OnChat(chat.Push)

	crack, err := NewCrack()
	if err != nil {
		return nil, err
	}

	spawn := bot.Snapshot()
	eye := eyeOf(spawn.Position)

	return &Viewer{
		window:   window,
		renderer: renderer,
		route:    route,
		hud:      hud,
		hand:     hand,
		chat:     chat,
		crack:    crack,
		edges:    newEdges(window),
		bot:      bot,
		from:     eye,
		to:       eye,
		lastTick: spawn.Tick,
		yaw:      spawn.Yaw,
		pitch:    spawn.Pitch,
		camera: Camera{
			Up:     mgl32.Vec3{0, 1, 0},
			FOV:    fieldOfView,
			Aspect: float32(defaultWidth) / float32(defaultHeight),
			Near:   nearPlane,
			Far:    farPlane,
		},
	}, nil
}

func (v *Viewer) Close() {
	v.crack.Close()
	v.hand.Close()
	v.hud.Close()
	v.route.Close()
	v.renderer.Close()
	v.window.Close()
}

func (v *Viewer) frame() {
	now := v.window.Time()

	v.follow()

	waypoints, next := v.bot.Route()
	v.route.Update(waypoints, next, v.bot.Snapshot().Position)
	v.hand.Update(v.bot.Inventory().Held())

	block, progress, digging := v.bot.Excavation()
	v.crack.Update(block, progress, digging)

	v.window.Clear(skyRed, skyGreen, skyBlue)
	v.renderer.Draw(v.camera)
	v.crack.Draw(v.camera)
	v.route.Draw(v.camera, now)
	v.hand.Draw(v.window, v.camera, now, v.swing(now))

	v.overlay(now)
}

func (v *Viewer) overlay(now float64) {
	screen := v.window.Viewport()

	v.window.Overlay(true)
	defer v.window.Overlay(false)

	v.hud.Draw(screen, v.bot.Inventory(), v.inventoryOpen)
	if v.inventoryOpen {
		v.hud.DrawCarried(v.bot.Carried(), v.window.Cursor())
	}

	v.chat.Draw(screen, now)
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
