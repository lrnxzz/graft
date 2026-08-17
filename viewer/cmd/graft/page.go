package main

import (
	"log/slog"

	"github.com/lrnxzz/graft/plugin"
	"github.com/lrnxzz/graft/viewer"
	"github.com/lrnxzz/graft/viewer/gpu"
	"github.com/lrnxzz/graft/viewer/ultralight"
)

// advances the html engine once a frame: there is one for the whole process
type repaint struct {
	engine *ultralight.Renderer
}

func (r repaint) Draw(*viewer.Canvas) {
	if r.engine == nil {
		return
	}

	r.engine.Update()
	r.engine.Render()
}

type paged struct {
	menu   *plugin.Menu
	page   *viewer.Page
	shown  string
	cursor gpu.Point
}

// built here and not on the first frame, so a menu opened and dismissed before
// it ever drew still has something to close
func newPaged(engine *ultralight.Renderer, screen gpu.Rect, menu *plugin.Menu, markup string) *paged {
	page := viewer.NewPage(engine, screen)

	opened := paged{
		menu:  menu,
		page:  page,
		shown: markup,
	}

	page.Sends(opened.send)
	page.Load(markup)

	return &opened
}

func (p *paged) Draw(canvas *viewer.Canvas) {
	// the wheel arrives without a position, so the pointer is remembered here;
	// hover and a scrolling box both depend on the page knowing where it is
	p.cursor = canvas.Cursor()
	p.page.Hover(p.cursor)

	p.refresh()
	p.page.Draw(canvas)
}

// refresh asks the plugin what the menu looks like now and reloads only when the
// answer changed; a reload costs a full relayout and loses scroll and hover
func (p *paged) refresh() {
	markup, err := document(p.menu)
	if err != nil {
		slog.Warn("plugin page", "err", err)

		return
	}
	if markup == "" || markup == p.shown {
		return
	}

	p.shown = markup
	p.page.Load(markup)
}

func document(menu *plugin.Menu) (string, error) {
	written, is, err := menu.Page()
	if err != nil || !is {
		return "", err
	}

	return viewer.Document(written.Style, written.Body)
}

func (p *paged) send(name string, args []string) {
	if err := p.menu.Send(name, args); err != nil {
		slog.Warn("plugin page", "message", name, "err", err)
	}
}

func (p *paged) Click(cursor gpu.Point, _ gpu.Rect) {
	p.page.Click(cursor)
}

func (p *paged) Scroll(delta float64) {
	p.page.Scroll(p.cursor, int(delta*scrollStep))
}

func (p *paged) Key(key gpu.Key) {
	p.page.Press(key)
}

func (p *paged) Close() {
	p.page.Close()
}
