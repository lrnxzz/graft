package main

import (
	"log/slog"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/plugin"
	"github.com/lrnxzz/graft/viewer"
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

type overlay struct {
	plugins *plugin.Plugins
}

func (o overlay) Draw(canvas *viewer.Canvas) {
	roots, failures := o.plugins.Hud()
	report("hud", failures)

	painted := surface{
		canvas: canvas,
		scale:  viewer.GuiScale,
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
	from := graft.Vec3(mark.From.X, mark.From.Y, mark.From.Z)
	to := graft.Vec3(mark.To.X, mark.To.Y, mark.To.Z)

	switch mark.Type {
	case plugin.MarkHighlight:
		painter.Highlight(from.Floor(), color)
	case plugin.MarkBeacon:
		painter.Line(from, from.Offset(0, pluginBeacon, 0), color)
	case plugin.MarkBox:
		painter.Box(graft.Box(from, to), color)
	case plugin.MarkLine:
		painter.Line(from, to, color)
	}
}

const pluginBeacon = 3

// a report is called from the frame loop, and a plugin broken there would say so
// sixty times a second — each distinct failure is said once
var reported = map[string]bool{}

func report(stage string, failures []plugin.Failure) {
	for _, failure := range failures {
		said := stage + ": " + failure.Error()
		if reported[said] {
			continue
		}
		reported[said] = true

		slog.Warn("plugin failed", "stage", stage, "err", failure)
	}
}
