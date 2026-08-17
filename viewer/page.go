package viewer

import (
	"github.com/lrnxzz/graft/viewer/gpu"
	"github.com/lrnxzz/graft/viewer/ultralight"
)

type Page struct {
	view    *ultralight.View
	surface *gpu.Surface
}

func NewPage(engine *ultralight.Renderer, screen gpu.Rect) *Page {
	width, height := int(screen.Width()), int(screen.Height())

	opened := Page{
		view:    engine.View(width, height),
		surface: gpu.NewSurface(width, height),
	}

	return &opened
}

// has to be set before anything loads: the bridge is built once per page
func (p *Page) Sends(handle func(name string, args []string)) {
	p.view.Sends(handle)
}

// what makes :hover work, and what decides which box the wheel moves
func (p *Page) Hover(cursor gpu.Point) {
	p.view.Moved(int(cursor.X), int(cursor.Y))
}

// a whole press: the viewer reports the edge, never the button coming back up
func (p *Page) Click(cursor gpu.Point) {
	x, y := int(cursor.X), int(cursor.Y)

	p.view.Moved(x, y)
	p.view.Pressed(x, y, ultralight.Left)
	p.view.Released(x, y, ultralight.Left)
}

// the document itself never scrolls, so markup that should move needs its own
// overflowing element
func (p *Page) Scroll(cursor gpu.Point, pixels int) {
	p.view.Moved(int(cursor.X), int(cursor.Y))
	p.view.Scrolled(0, pixels)
}

func (p *Page) Press(key gpu.Key) {
	code := ultralight.Virtual(int(key))

	p.view.KeyDown(code, 0)
	p.view.KeyUp(code, 0)
}

func (p *Page) Load(html string) {
	p.view.LoadHTML(html)
}

// resizing drops the texture's contents, so the engine marks the whole page dirty
func (p *Page) fit(screen gpu.Rect) {
	width, height := int(screen.Width()), int(screen.Height())

	wide, tall := p.surface.Size()
	if width == wide && height == tall {
		return
	}

	p.view.Resize(width, height)
	p.surface.Resize(width, height)
}

// The engine is advanced by the frame, not here: there is one for the whole
// process, so a page that ticked it would tick it once per page on screen.
func (p *Page) Draw(canvas *Canvas) {
	p.fit(canvas.Screen())
	p.upload()

	width, height := p.surface.Size()
	canvas.Page(p.surface, gpu.RectAt(gpu.Point{}, float32(width), float32(height)))
}

// upload moves only the rows the page repainted, which is what keeps a blinking
// cursor from costing a megabyte a frame
func (p *Page) upload() {
	changed := p.view.Dirty()
	if changed.Empty() {
		return
	}

	p.view.Pixels(func(pixels []byte, stride int) {
		p.surface.Upload(pixels, stride, changed.Left, changed.Top, changed.Width(), changed.Height())
	})

	p.view.Clean()
}

func (p *Page) Close() {
	p.view.Close()
	p.surface.Delete()
}
