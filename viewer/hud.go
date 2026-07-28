package viewer

import (
	"bytes"
	_ "embed"
	"image/png"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

//go:embed assets/shaders/hud.vert
var hudVertexShader string

//go:embed assets/shaders/hud.frag
var hudFragmentShader string

//go:embed assets/hotbar.png
var hotbarImage []byte

//go:embed assets/hotbar_selection.png
var selectionImage []byte

//go:embed assets/inventory.png
var containerImage []byte

//go:embed assets/crosshair.png
var crosshairImage []byte

const (
	guiScale = 3

	hotbarWidth     = 182
	hotbarHeight    = 22
	hotbarStride    = 20
	hotbarInset     = 3
	selectionWidth  = 24
	selectionHeight = 23

	containerWidth  = 176
	containerHeight = 166
	containerCanvas = 256

	crosshairSize = 15

	slotSize = 18
	iconSize = 16

	hudFloatsPerVertex = 8
	hudFloatsPerQuad   = 4 * hudFloatsPerVertex
)

var wholeTexture = gpu.UV{
	U0: 0,
	V0: 0,
	U1: 1,
	V1: 1,
}

type hudColor struct {
	red   float32
	green float32
	blue  float32
	alpha float32
}

func shade(red, green, blue, alpha float32) hudColor {
	return hudColor{
		red:   red,
		green: green,
		blue:  blue,
		alpha: alpha,
	}
}

var opaqueWhite = shade(1, 1, 1, 1)

type hudMeshes struct {
	bar      *gpu.Mesh
	frame    *gpu.Mesh
	slots    *gpu.Mesh
	panel    *gpu.Mesh
	contents *gpu.Mesh
	sight    *gpu.Mesh
}

func (m *hudMeshes) release() {
	uploaded := [...]*gpu.Mesh{
		m.bar,
		m.frame,
		m.slots,
		m.panel,
		m.contents,
		m.sight,
	}
	for _, mesh := range uploaded {
		if mesh != nil {
			mesh.Delete()
		}
	}

	*m = hudMeshes{}
}

type Hud struct {
	program   *gpu.Program
	icons     *Iconset
	hotbar    *gpu.Texture
	selection *gpu.Texture
	container *gpu.Texture
	crosshair *gpu.Texture

	meshes hudMeshes

	items  [gocraft.InventorySize]gocraft.ItemID
	held   int
	screen gpu.Rect
}

func NewHud(icons *Iconset) (*Hud, error) {
	program, err := gpu.NewProgram(hudVertexShader, hudFragmentShader)
	if err != nil {
		return nil, err
	}

	hotbar, err := loadTexture(hotbarImage)
	if err != nil {
		return nil, err
	}
	selection, err := loadTexture(selectionImage)
	if err != nil {
		return nil, err
	}
	container, err := loadTexture(containerImage)
	if err != nil {
		return nil, err
	}
	crosshair, err := loadTexture(crosshairImage)
	if err != nil {
		return nil, err
	}

	return &Hud{
		program:   program,
		icons:     icons,
		hotbar:    hotbar,
		selection: selection,
		container: container,
		crosshair: crosshair,
		held:      -1,
	}, nil
}

func loadTexture(encoded []byte) (*gpu.Texture, error) {
	img, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}

	return gpu.NewTexture(img), nil
}

func (h *Hud) Draw(screen gpu.Rect, inventory *gocraft.Inventory, open bool) {
	var items [gocraft.InventorySize]gocraft.ItemID
	for index := range gocraft.InventorySize {
		items[index] = inventory.Slot(index).Item
	}

	held := inventory.HeldIndex()
	if items != h.items || held != h.held || screen != h.screen {
		h.items = items
		h.held = held
		h.screen = screen
		h.rebuild()
	}

	h.bind()

	h.hotbar.Bind(0)
	h.meshes.bar.Draw()
	h.selection.Bind(0)
	h.meshes.frame.Draw()
	h.icons.Atlas().Bind(0)
	h.meshes.slots.Draw()

	if !open {
		h.crosshair.Bind(0)
		h.meshes.sight.Draw()

		return
	}

	h.container.Bind(0)
	h.meshes.panel.Draw()
	h.icons.Atlas().Bind(0)
	h.meshes.contents.Draw()
}

func (h *Hud) bind() {
	h.program.Use()
	h.program.Vec2("screen", h.screen.Width(), h.screen.Height())
	h.program.Int("icons", 0)
}

func (h *Hud) Close() {
	h.meshes.release()
}

func (h *Hud) SlotAt(cursor gpu.Point, screen gpu.Rect) (int, bool) {
	panel := screen.Centered(containerWidth*guiScale, containerHeight*guiScale)

	for index := range gocraft.InventorySize {
		if slotCell(panel, containerSlot(index)).Contains(cursor) {
			return index, true
		}
	}

	return 0, false
}

// the clickable cell is the 18px slot, whose art starts one pixel above and to
// the left of the 16px icon the layout points at
func slotCell(panel gpu.Rect, origin gpu.Point) gpu.Rect {
	corner := panel.Min.Add(origin.Offset(-1, -1).Scale(guiScale))

	return gpu.RectAt(corner, slotSize*guiScale, slotSize*guiScale)
}

func iconAt(origin gpu.Point) gpu.Rect {
	return gpu.RectAt(origin, iconSize*guiScale, iconSize*guiScale)
}

func (h *Hud) DrawCarried(carried gocraft.ItemStack, cursor gpu.Point) {
	uv, drawable := h.icons.Icon(carried.Item)
	if carried.Empty() || !drawable {
		return
	}

	h.bind()
	h.icons.Atlas().Bind(0)

	icon := iconAt(cursor.Offset(-iconSize*guiScale/2, -iconSize*guiScale/2))

	held := uploadHud(hudQuad(nil, icon, uv))
	held.Draw()
	held.Delete()
}

var containerSlots = map[int]gpu.Point{
	gocraft.SlotCraftingOutput: gpu.At(154, 28),
	gocraft.SlotHead:           gpu.At(8, 8),
	gocraft.SlotChest:          gpu.At(8, 26),
	gocraft.SlotLegs:           gpu.At(8, 44),
	gocraft.SlotFeet:           gpu.At(8, 62),
	gocraft.SlotOffhand:        gpu.At(77, 62),
}

func craftingSlot(index int) gpu.Point {
	grid := index - 1

	return gpu.At(98+float32(grid%2*slotSize), 18+float32(grid/2*slotSize))
}

func mainSlot(index int) gpu.Point {
	row := (index - gocraft.SlotMainStart) / gocraft.HotbarSize
	column := (index - gocraft.SlotMainStart) % gocraft.HotbarSize

	return gpu.At(8+float32(column*slotSize), 84+float32(row*slotSize))
}

func hotbarSlot(index int) gpu.Point {
	return gpu.At(8+float32((index-gocraft.SlotHotbarStart)*slotSize), 142)
}

func containerSlot(index int) gpu.Point {
	switch {
	case index >= gocraft.SlotHotbarStart && index < gocraft.SlotOffhand:
		return hotbarSlot(index)
	case index >= gocraft.SlotMainStart && index < gocraft.SlotHotbarStart:
		return mainSlot(index)
	case index >= 1 && index < gocraft.SlotHead:
		return craftingSlot(index)
	default:
		return containerSlots[index]
	}
}

func (h *Hud) rebuild() {
	h.meshes.release()

	const barWidth, barHeight = hotbarWidth * guiScale, hotbarHeight * guiScale

	// the hotbar is centred across the screen and flush with its bottom edge
	bar := gpu.RectAt(
		gpu.At(h.screen.Center().X-barWidth/2, h.screen.Max.Y-barHeight),
		barWidth, barHeight)
	h.meshes.bar = uploadHud(hudQuad(nil, bar, wholeTexture))

	frame := gpu.RectAt(
		bar.Min.Offset(float32(h.held*hotbarStride-1)*guiScale, -guiScale),
		selectionWidth*guiScale, selectionHeight*guiScale)
	h.meshes.frame = uploadHud(hudQuad(nil, frame, wholeTexture))

	sight := h.screen.Centered(crosshairSize*guiScale, crosshairSize*guiScale)
	h.meshes.sight = uploadHud(hudQuad(nil, sight, wholeTexture))

	panel := h.screen.Centered(containerWidth*guiScale, containerHeight*guiScale)
	art := gpu.UV{
		U0: 0,
		V0: 0,
		U1: float32(containerWidth) / containerCanvas,
		V1: float32(containerHeight) / containerCanvas,
	}
	h.meshes.panel = uploadHud(hudQuad(nil, panel, art))

	h.meshes.slots = uploadHud(h.hotbarIcons(bar))
	h.meshes.contents = uploadHud(h.containerIcons(panel))
}

func (h *Hud) hotbarIcons(bar gpu.Rect) []float32 {
	var vertices []float32
	for index := range gocraft.HotbarSize {
		uv, drawable := h.icons.Icon(h.items[gocraft.HotbarSlot(index)])
		if !drawable {
			continue
		}

		origin := bar.Min.Offset(float32(hotbarInset+index*hotbarStride)*guiScale, hotbarInset*guiScale)
		vertices = hudQuad(vertices, iconAt(origin), uv)
	}

	return vertices
}

func (h *Hud) containerIcons(panel gpu.Rect) []float32 {
	var vertices []float32
	for index := range gocraft.InventorySize {
		uv, drawable := h.icons.Icon(h.items[index])
		if !drawable {
			continue
		}

		origin := panel.Min.Add(containerSlot(index).Scale(guiScale))
		vertices = hudQuad(vertices, iconAt(origin), uv)
	}

	return vertices
}

func hudQuad(vertices []float32, area gpu.Rect, uv gpu.UV) []float32 {
	return hudTintedQuad(vertices, area, uv, opaqueWhite)
}

func hudTintedQuad(vertices []float32, area gpu.Rect, uv gpu.UV, tint hudColor) []float32 {
	red, green, blue, alpha := tint.red, tint.green, tint.blue, tint.alpha

	return append(vertices,
		area.Min.X, area.Max.Y, uv.U0, uv.V1, red, green, blue, alpha,
		area.Max.X, area.Max.Y, uv.U1, uv.V1, red, green, blue, alpha,
		area.Max.X, area.Min.Y, uv.U1, uv.V0, red, green, blue, alpha,
		area.Min.X, area.Min.Y, uv.U0, uv.V0, red, green, blue, alpha)
}

func uploadHud(vertices []float32) *gpu.Mesh {
	return gpu.NewMesh(vertices, gpu.QuadIndices(len(vertices)/hudFloatsPerQuad),
		gpu.Attribute{Location: 0, Size: 2},
		gpu.Attribute{Location: 1, Size: 2},
		gpu.Attribute{Location: 2, Size: 4})
}
