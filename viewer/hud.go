package viewer

import (
	"bytes"
	_ "embed"
	"image/png"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/agent"
	"github.com/lrnxzz/graft/viewer/gpu"
)

//go:embed assets/hotbar.png
var hotbarImage []byte

//go:embed assets/hotbar_selection.png
var selectionImage []byte

//go:embed assets/crosshair.png
var crosshairImage []byte

const (
	GuiScale = 3

	hotbarWidth     = 182
	hotbarHeight    = 22
	hotbarStride    = 20
	hotbarInset     = 3
	selectionWidth  = 24
	selectionHeight = 23

	crosshairSize = 15

	slotSize = 18
	iconSize = 16
)

var wholeTexture = gpu.UV{
	U0: 0,
	V0: 0,
	U1: 1,
	V1: 1,
}

// Hud is the always-on Layer: the hotbar and the crosshair. The inventory panel
// is deliberately not here — it takes input, so it is a Screen.
type Hud struct {
	bot       *agent.Agent
	hotbar    *gpu.Texture
	selection *gpu.Texture
	crosshair *gpu.Texture

	// vanilla hides the crosshair while a menu is up, and the viewer owns that fact
	aiming bool
}

func NewHud(bot *agent.Agent) (*Hud, error) {
	hotbar, err := loadTexture(hotbarImage)
	if err != nil {
		return nil, err
	}
	selection, err := loadTexture(selectionImage)
	if err != nil {
		return nil, err
	}
	crosshair, err := loadTexture(crosshairImage)
	if err != nil {
		return nil, err
	}

	return &Hud{
		bot:       bot,
		hotbar:    hotbar,
		selection: selection,
		crosshair: crosshair,
		aiming:    true,
	}, nil
}

func loadTexture(encoded []byte) (*gpu.Texture, error) {
	img, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}

	return gpu.NewTexture(img), nil
}

func (h *Hud) Draw(canvas *Canvas) {
	const barWidth, barHeight = hotbarWidth * GuiScale, hotbarHeight * GuiScale

	screen := canvas.Screen()
	inventory := h.bot.Inventory()

	// the hotbar is centred across the screen and flush with its bottom edge
	bar := gpu.RectAt(
		gpu.At(screen.Center().X-barWidth/2, screen.Max.Y-barHeight),
		barWidth, barHeight)
	canvas.Sprite(h.hotbar, bar, wholeTexture, gpu.White)

	frame := gpu.RectAt(
		bar.Min.Offset(float32(inventory.HeldIndex()*hotbarStride-1)*GuiScale, -GuiScale),
		selectionWidth*GuiScale, selectionHeight*GuiScale)
	canvas.Sprite(h.selection, frame, wholeTexture, gpu.White)

	for index := range graft.HotbarSize {
		origin := bar.Min.Offset(float32(hotbarInset+index*hotbarStride)*GuiScale, hotbarInset*GuiScale)
		canvas.Icon(inventory.Hotbar(index).Item, iconAt(origin))
	}

	if h.aiming {
		sight := screen.Centered(crosshairSize*GuiScale, crosshairSize*GuiScale)
		canvas.Sprite(h.crosshair, sight, wholeTexture, gpu.White)
	}
}

func (h *Hud) Close() {
	for _, texture := range [...]*gpu.Texture{h.hotbar, h.selection, h.crosshair} {
		texture.Delete()
	}
}

func iconAt(origin gpu.Point) gpu.Rect {
	return gpu.RectAt(origin, iconSize*GuiScale, iconSize*GuiScale)
}

// the crosshair is hidden while a menu is up
func (h *Hud) Aiming(at bool) {
	h.aiming = at
}
