package viewer

import (
	_ "embed"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

//go:embed assets/inventory.png
var containerImage []byte

const (
	containerWidth  = 176
	containerHeight = 166
	containerCanvas = 256
)

// InventoryScreen is the vanilla container panel, and the first Screen the viewer
// ships. It is written against the same API a plugin gets, so if a menu cannot be
// built this way the API is wrong.
type InventoryScreen struct {
	bot   *agent.Agent
	panel *gpu.Texture
}

func NewInventoryScreen(bot *agent.Agent) (*InventoryScreen, error) {
	panel, err := loadTexture(containerImage)
	if err != nil {
		return nil, err
	}

	return &InventoryScreen{
		bot:   bot,
		panel: panel,
	}, nil
}

func (s *InventoryScreen) Draw(canvas *Canvas) {
	frame := containerFrame(canvas.Screen())

	art := gpu.UV{
		U0: 0,
		V0: 0,
		U1: float32(containerWidth) / containerCanvas,
		V1: float32(containerHeight) / containerCanvas,
	}
	canvas.Sprite(s.panel, frame, art, gpu.White)

	inventory := s.bot.Inventory()
	for index := range gocraft.InventorySize {
		origin := frame.Min.Add(containerSlot(index).Scale(guiScale))
		canvas.Icon(inventory.Slot(index).Item, iconAt(origin))
	}

	s.drawCarried(canvas)
}

// the stack on the cursor rides above the panel, centred on the pointer
func (s *InventoryScreen) drawCarried(canvas *Canvas) {
	carried := s.bot.Carried()
	if carried.Empty() {
		return
	}

	cursor := canvas.Cursor()
	canvas.Icon(carried.Item, iconAt(cursor.Offset(-iconSize*guiScale/2, -iconSize*guiScale/2)))
}

func (s *InventoryScreen) Click(cursor gpu.Point, screen gpu.Rect) {
	slot, hit := slotAt(cursor, screen)
	if !hit {
		return
	}

	_ = s.bot.ClickSlot(slot)
}

func (s *InventoryScreen) Key(gpu.Key) {
}

func (s *InventoryScreen) Close() {
	s.panel.Delete()
}

func slotAt(cursor gpu.Point, screen gpu.Rect) (int, bool) {
	frame := containerFrame(screen)

	for index := range gocraft.InventorySize {
		if slotCell(frame, containerSlot(index)).Contains(cursor) {
			return index, true
		}
	}

	return 0, false
}

func containerFrame(screen gpu.Rect) gpu.Rect {
	return screen.Centered(containerWidth*guiScale, containerHeight*guiScale)
}

// the clickable cell is the 18px slot, whose art starts one pixel above and to
// the left of the 16px icon the layout points at
func slotCell(panel gpu.Rect, origin gpu.Point) gpu.Rect {
	corner := panel.Min.Add(origin.Offset(-1, -1).Scale(guiScale))

	return gpu.RectAt(corner, slotSize*guiScale, slotSize*guiScale)
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
