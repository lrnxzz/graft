package plugin

import (
	"errors"

	"github.com/dop251/goja"
)

var errNotANode = errors.New("a hud or menu body returned something that is not a node")

type Menu struct {
	Title   string
	runtime *Runtime
	body    goja.Callable
	page    goja.Callable
	style   string
	on      map[string]goja.Callable
	onKey   goja.Callable
}

func (m *Menu) Body() (Node, bool, error) {
	if m.body == nil {
		return nil, false, nil
	}

	built, err := m.body(goja.Undefined(), m.runtime.vm.Get("bot"))
	if err != nil {
		return nil, false, err
	}

	return rootOf(m.runtime.reading(built))
}

// Markup is what a plugin wrote when it describes a menu as a page rather than
// as a tree of nodes. It is a fragment and a stylesheet, not a document: the
// host assembles that, since only the host knows which sprites and fonts the
// page needs and how the window wants to be drawn on.
type Markup struct {
	Body  string
	Style string
}

// Page reports how the menu was written. A menu declares a page or a body, and
// this answers whether it was the first.
func (m *Menu) Page() (Markup, bool, error) {
	if m.page == nil {
		return Markup{}, false, nil
	}

	built, err := m.page(goja.Undefined(), m.runtime.vm.Get("bot"))
	if err != nil {
		return Markup{}, false, err
	}

	written := m.runtime.reading(built)
	if written.missing() {
		return Markup{}, false, nil
	}

	return Markup{
		Body:  written.text(),
		Style: m.style,
	}, true, nil
}

// Written reports whether the menu is a page rather than a tree of nodes. The
// two are alternatives, and the host has to choose a renderer before it asks
// the plugin to build anything — asking Page() to find out would run the
// plugin's callback on the path that then throws the answer away.
func (m *Menu) Written() bool {
	return m.page != nil
}

// Send delivers what the page asked for through graft.send. An unclaimed name
// is not an error: markup outlives the handler that used to answer it.
func (m *Menu) Send(name string, args []string) error {
	handle := m.on[name]
	if handle == nil {
		return nil
	}

	passed := make([]goja.Value, 0, len(args)+1)
	for _, arg := range args {
		passed = append(passed, m.runtime.vm.ToValue(arg))
	}
	passed = append(passed, m.runtime.vm.Get("bot"))

	_, err := handle(goja.Undefined(), passed...)

	return err
}

func (m *Menu) Key(key string) error {
	if m.onKey == nil {
		return nil
	}

	_, err := m.onKey(goja.Undefined(), m.runtime.vm.ToValue(key), m.runtime.vm.Get("bot"))

	return err
}

type UI interface {
	Open(menu *Menu)
	Dismiss()
	Showing() bool
	Toast(text string)
}

type capturing struct {
	inner UI
	menu  *Menu
}

func (c *capturing) Open(menu *Menu) {
	c.menu = menu
}

func (c *capturing) Dismiss() {
	if c.inner != nil {
		c.inner.Dismiss()
	}
}

func (c *capturing) Showing() bool {
	return c.inner != nil && c.inner.Showing()
}

func (c *capturing) Toast(text string) {
	if c.inner != nil {
		c.inner.Toast(text)
	}
}

func (r *Runtime) uiObject(ui UI) *goja.Object {
	object := r.vm.NewObject()

	_ = into(r.vm, object).all(map[string]any{
		"open": func(declared goja.Value) {
			if ui == nil {
				return
			}

			spec := r.reading(declared)
			opened := Menu{
				Title:   spec.field("title").text(),
				runtime: r,
				body:    spec.field("body").callable(),
				page:    spec.field("page").callable(),
				style:   spec.field("style").text(),
				on:      handlers(spec.field("on")),
				onKey:   spec.field("onKey").callable(),
			}

			ui.Open(&opened)
		},
		"dismiss": func() {
			if ui != nil {
				ui.Dismiss()
			}
		},
		"toast": func(line string) {
			if ui != nil {
				ui.Toast(line)
			}
		},
	}).getter("showing", func() any {
		return ui != nil && ui.Showing()
	}).done()

	return object
}

func handlers(given reading) map[string]goja.Callable {
	names := given.keys()
	if len(names) == 0 {
		return nil
	}

	claimed := make(map[string]goja.Callable, len(names))
	for _, name := range names {
		handle := given.field(name).callable()
		if handle != nil {
			claimed[name] = handle
		}
	}

	return claimed
}

func rootOf(given reading) (Node, bool, error) {
	if given.missing() {
		return nil, false, nil
	}

	node, built := nodeOf(given)
	if !built {
		return nil, false, errNotANode
	}

	return node, true, nil
}
