package main

import (
	"log/slog"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/plugin"
	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const guiScale = 3

type overlay struct {
	plugins *plugin.Plugins
}

func (o overlay) Draw(canvas *viewer.Canvas) {
	roots, failures := o.plugins.Hud()
	report("hud", failures)

	painted := surface{
		canvas: canvas,
		scale:  guiScale,
	}
	for _, root := range roots {
		plugin.Paint(root, painted)
	}
}

type pulse struct {
	plugins *plugin.Plugins
}

func (p pulse) Draw(*viewer.Canvas) {
	report("tick", p.plugins.Tick())
}

type marks struct {
	plugins *plugin.Plugins
}

func (m marks) DrawWorld(painter *viewer.Painter) {
	drawn, failures := m.plugins.World()
	report("world", failures)

	for _, mark := range drawn {
		paint(painter, mark)
	}
}

func paint(painter *viewer.Painter, mark plugin.Marker) {
	color := tint(mark.Color)
	from := gocraft.Vec3(mark.From.X, mark.From.Y, mark.From.Z)
	to := gocraft.Vec3(mark.To.X, mark.To.Y, mark.To.Z)

	switch mark.Kind {
	case plugin.MarkHighlight:
		painter.Highlight(from.Floor(), color)
	case plugin.MarkBeacon:
		painter.Line(from, from.Offset(0, beaconHeight, 0), color)
	case plugin.MarkBox:
		painter.Box(gocraft.Box(from, to), color)
	case plugin.MarkLine:
		painter.Line(from, to, color)
	}
}

const beaconHeight = 3

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
		scale:  guiScale,
	}
	s.picks = plugin.Scrolled(root, painted, &s.scroll)
}

func (s *screen) Click(cursor gpu.Point, _ gpu.Rect) {
	hit := plugin.Clicked(s.picks, cursor.X/guiScale, cursor.Y/guiScale)
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

func report(stage string, failures []plugin.Failure) {
	for _, failure := range failures {
		slog.Warn("plugin "+stage, "err", failure)
	}
}
