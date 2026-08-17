package main

import (
	"log/slog"

	"github.com/lrnxzz/graft/plugin"
	"github.com/lrnxzz/graft/viewer"
	"github.com/lrnxzz/graft/viewer/gpu"
)

// A screen is a menu drawn as a node tree on the canvas, which is the shape a
// plugin declares when it never asked for html.
type screen struct {
	menu   *plugin.Menu
	picks  []plugin.Pick
	scroll plugin.Scroll
}

func (s *screen) Draw(canvas *viewer.Canvas) {
	root, built, err := s.menu.Body()
	if err != nil {
		slog.Warn("plugin menu", "err", err)

		return
	}
	if !built {
		return
	}

	painted := surface{
		canvas: canvas,
		scale:  viewer.GuiScale,
	}
	s.picks = plugin.Scrolled(root, painted, &s.scroll)
}

func (s *screen) Click(cursor gpu.Point, _ gpu.Rect) {
	hit := plugin.Clicked(s.picks, cursor.X/viewer.GuiScale, cursor.Y/viewer.GuiScale)
	if hit != nil {
		hit()
	}
}

func (s *screen) Key(gpu.Key) {
}

func (s *screen) Scroll(delta float64) {
	s.scroll.By(float32(delta) * scrollStep)
}

const scrollStep = 12
