package main

import (
	"log/slog"

	"github.com/lrnxzz/go-craft/plugin"
	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

// repaint advances the html engine, once, at the top of every frame. There is
// one engine for the whole process, so leaving this to the pages would run it
// again for each one on screen.
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

// paged is the other kind of plugin menu: HTML and CSS drawn by the engine
// rather than a tree of nodes drawn by the canvas. Clicks and the wheel go to
// the document, and what the document sends comes back to the plugin.
type paged struct {
	menu   *plugin.Menu
	page   *viewer.Page
	shown  string
	cursor gpu.Point
}

// the page is built here rather than on the first frame, so that a menu opened
// and dismissed before it ever drew still has something to close
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

// document is what the engine loads: the plugin says what the menu is, the
// viewer says how a page is drawn and which sprites it needs
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
