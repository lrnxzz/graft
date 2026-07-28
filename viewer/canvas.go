package viewer

import (
	_ "embed"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

//go:embed assets/shaders/hud.vert
var hudVertexShader string

//go:embed assets/shaders/hud.frag
var hudFragmentShader string

// the hud shader reads a negative u as "no texture", which is what lets a flat
// rectangle share a batch with whatever is already bound
var untextured = gpu.UV{
	U0: -1,
	V0: -1,
	U1: -1,
	V1: -1,
}

const (
	hudFloatsPerVertex = 8
	hudFloatsPerQuad   = 4 * hudFloatsPerVertex
)

// a batch is one draw call: the texture bound for it and the quads using it
type batch struct {
	texture  *gpu.Texture
	vertices []float32
}

// Canvas is the 2D surface a Layer or a Screen paints on. It hides the vertex
// format, the atlases and the batching, so a plugin only ever names rectangles,
// text and items.
type Canvas struct {
	screen  gpu.Rect
	cursor  gpu.Point
	time    float64
	font    *Font
	icons   *Iconset
	batches []batch
}

// Screen is the whole drawable area, the anchor every layout is derived from
func (c *Canvas) Screen() gpu.Rect {
	return c.screen
}

// Cursor is where the pointer sits, so a screen can draw what is under it
func (c *Canvas) Cursor() gpu.Point {
	return c.cursor
}

// Fill paints a flat rectangle
func (c *Canvas) Fill(area gpu.Rect, tint gpu.Color) {
	c.quad(nil, area, untextured, tint)
}

// Sprite paints a piece of a texture
func (c *Canvas) Sprite(texture *gpu.Texture, area gpu.Rect, uv gpu.UV, tint gpu.Color) {
	c.quad(texture, area, uv, tint)
}

// Icon paints an item's inventory sprite and reports whether the item has one
func (c *Canvas) Icon(item gocraft.ItemID, area gpu.Rect) bool {
	uv, drawable := c.icons.Icon(item)
	if !drawable {
		return false
	}

	c.quad(c.icons.Atlas().Texture(), area, uv, gpu.White)

	return true
}

// Text paints one line and answers where the pen came to rest, so callers can
// keep writing without measuring again
func (c *Canvas) Text(text string, pen gpu.Point, scale float32, tint gpu.Color) gpu.Point {
	for _, letter := range text {
		uv, advance, drawable := c.font.glyph(letter)
		if !drawable {
			continue
		}

		size := glyphSize * scale
		c.quad(c.font.texture, gpu.RectAt(pen, size, size), uv, tint)
		pen = pen.Offset(advance*scale, 0)
	}

	return pen
}

// Shadowed is the vanilla look for text: the same line twice, once dark and
// offset by a pixel of the current scale
func (c *Canvas) Shadowed(text string, pen gpu.Point, scale float32, tint gpu.Color) gpu.Point {
	c.Text(text, pen.Offset(scale, scale), scale, textShadow)

	return c.Text(text, pen, scale, tint)
}

func (c *Canvas) TextWidth(text string, scale float32) float32 {
	return c.font.Width(text) * scale
}

func (c *Canvas) quad(texture *gpu.Texture, area gpu.Rect, uv gpu.UV, tint gpu.Color) {
	target := c.open(texture)
	target.vertices = append(target.vertices,
		area.Min.X, area.Max.Y, uv.U0, uv.V1, tint.Red, tint.Green, tint.Blue, tint.Alpha,
		area.Max.X, area.Max.Y, uv.U1, uv.V1, tint.Red, tint.Green, tint.Blue, tint.Alpha,
		area.Max.X, area.Min.Y, uv.U1, uv.V0, tint.Red, tint.Green, tint.Blue, tint.Alpha,
		area.Min.X, area.Min.Y, uv.U0, uv.V0, tint.Red, tint.Green, tint.Blue, tint.Alpha)
}

// open answers the batch new geometry belongs to, starting a fresh one only when
// the texture really changes; untextured quads settle for whatever is open
func (c *Canvas) open(texture *gpu.Texture) *batch {
	if len(c.batches) > 0 {
		last := &c.batches[len(c.batches)-1]

		switch {
		case texture == nil || last.texture == texture:
			return last
		case last.texture == nil:
			last.texture = texture

			return last
		}
	}

	c.batches = append(c.batches, batch{texture: texture})

	return &c.batches[len(c.batches)-1]
}

// Time is the running clock of the window, what an animated overlay advances on
func (c *Canvas) Time() float64 {
	return c.time
}

func (c *Canvas) reset(screen gpu.Rect, cursor gpu.Point, now float64) {
	c.screen = screen
	c.cursor = cursor
	c.time = now
	c.batches = c.batches[:0]
}

// paint uploads each batch in the order it was requested; overlay geometry is a
// few dozen quads a frame, so it is rebuilt rather than cached
func (c *Canvas) paint(program *gpu.Program) {
	program.Use()
	program.Vec2("screen", c.screen.Width(), c.screen.Height())
	program.Int("icons", 0)

	for _, drawn := range c.batches {
		if len(drawn.vertices) == 0 {
			continue
		}
		if drawn.texture != nil {
			drawn.texture.Bind(0)
		}

		mesh := gpu.NewMesh(drawn.vertices, gpu.QuadIndices(len(drawn.vertices)/hudFloatsPerQuad),
			gpu.Attribute{Location: 0, Size: 2},
			gpu.Attribute{Location: 1, Size: 2},
			gpu.Attribute{Location: 2, Size: 4})
		mesh.Draw()
		mesh.Delete()
	}
}
