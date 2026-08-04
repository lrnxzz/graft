package viewer

import (
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

// Page is HTML drawn over the world. It owns the engine's view and the texture
// the frame reads, and keeps the two the same size as the window, so a caller
// writes markup and never touches either.
type Page struct {
	view    *ultralight.View
	surface *gpu.Surface
}

// NewPage opens a page filling the given area. The engine has to be the one the
// viewer opened, since there is only ever one per process.
func NewPage(engine *ultralight.Renderer, screen gpu.Rect) *Page {
	width, height := int(screen.Width()), int(screen.Height())

	opened := Page{
		view:    engine.View(width, height),
		surface: gpu.NewSurface(width, height),
	}

	return &opened
}

// Sends decides what happens when the page calls gocraft.send. It has to be set
// before anything is loaded, since the bridge is built once per page.
func (p *Page) Sends(handle func(name string, args []string)) {
	p.view.Sends(handle)
}

// Hover keeps the document's idea of the pointer current, which is what makes
// :hover work and what decides which box the wheel moves
func (p *Page) Hover(cursor gpu.Point) {
	p.view.Moved(int(cursor.X), int(cursor.Y))
}

// Click delivers a whole press, since the viewer reports the edge rather than
// the button coming back up
func (p *Page) Click(cursor gpu.Point) {
	x, y := int(cursor.X), int(cursor.Y)

	p.view.Moved(x, y)
	p.view.Pressed(x, y, ultralight.Left)
	p.view.Released(x, y, ultralight.Left)
}

// Scroll moves the box under the pointer. The document itself never scrolls, so
// markup that should move needs its own overflowing element.
func (p *Page) Scroll(cursor gpu.Point, pixels int) {
	p.view.Moved(int(cursor.X), int(cursor.Y))
	p.view.Scrolled(0, pixels)
}

// Press sends a whole stroke. The window and the engine agree on letters, digits
// and space; everything else is named, and the naming lives with the engine
// rather than with whoever happens to be holding a key.
func (p *Page) Press(key gpu.Key) {
	code := ultralight.Virtual(int(key))

	p.view.KeyDown(code, 0)
	p.view.KeyUp(code, 0)
}

func (p *Page) Load(html string) {
	p.view.LoadHTML(html)
}

// fit follows the window. Resizing drops the texture's contents, so the engine
// marks the whole page dirty rather than the region it would otherwise report.
func (p *Page) fit(screen gpu.Rect) {
	width, height := int(screen.Width()), int(screen.Height())

	wide, tall := p.surface.Size()
	if width == wide && height == tall {
		return
	}

	p.view.Resize(width, height)
	p.surface.Resize(width, height)
}

// Draw uploads whatever the engine repainted. It never waits for a load to
// finish: an unloaded page simply has nothing dirty yet.
//
// Advancing the engine is not done here. There is one engine for the whole
// process, so a page that ticked it would tick it again for every other page on
// screen; that belongs to the frame, in Advance.
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
