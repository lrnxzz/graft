package viewer_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/viewer"
	"github.com/lrnxzz/graft/viewer/gpu"
)

// The three types below are the plugin the API promises is possible: one marking
// up the world, one painting an overlay and one menu. They live in an external
// package on purpose — anything they cannot reach is genuinely out of a plugin's
// reach.

type waypointMarker struct {
	at graft.Position
}

func (m waypointMarker) DrawWorld(painter *viewer.Painter) {
	painter.Highlight(m.at, gpu.RGBA(1, 0.8, 0, 0.9))
	painter.Line(m.at.Center(), m.at.Center().Offset(0, 3, 0), gpu.White)
	painter.Box(graft.Box(m.at.Corner(), m.at.Corner().Offset(1, 1, 1)), gpu.White)
}

type statusPanel struct {
	title string
}

func (p statusPanel) Draw(canvas *viewer.Canvas) {
	const margin = 8

	frame := gpu.RectAt(
		canvas.Screen().Min.Offset(margin, margin),
		canvas.TextWidth(p.title, 2)+2*margin,
		32)

	canvas.Fill(frame, gpu.RGBA(0, 0, 0, 0.5))
	canvas.Shadowed(p.title, frame.Min.Offset(margin, margin), 2, gpu.White)
}

type pickerScreen struct {
	items  []graft.ItemID
	chosen graft.ItemID
}

const pickerCell = 48

func (s *pickerScreen) Draw(canvas *viewer.Canvas) {
	for index, item := range s.items {
		canvas.Icon(item, s.cell(index, canvas.Screen()))
	}
}

func (s *pickerScreen) Click(cursor gpu.Point, screen gpu.Rect) {
	for index, item := range s.items {
		if s.cell(index, screen).Contains(cursor) {
			s.chosen = item

			return
		}
	}
}

func (s *pickerScreen) Key(gpu.Key) {
}

func (s *pickerScreen) cell(index int, screen gpu.Rect) gpu.Rect {
	origin := screen.Center().Offset(float32(index)*pickerCell, 0)

	return gpu.RectAt(origin, pickerCell, pickerCell)
}

// the seams are proven at compile time; nothing about them is worth asserting at
// runtime, and the build failing is the louder signal anyway
var (
	_ viewer.WorldLayer = waypointMarker{}
	_ viewer.Layer      = statusPanel{}
	_ viewer.Screen     = (*pickerScreen)(nil)
)

// a menu is only usable if a click can be resolved back to what was drawn, so the
// hit test is exercised against the same layout Draw uses
func TestPluginScreenResolvesAClickToTheItemUnderIt(t *testing.T) {
	screen := gpu.RectAt(gpu.At(0, 0), 800, 600)
	picker := &pickerScreen{items: []graft.ItemID{11, 22, 33}}

	second := picker.cell(1, screen)
	picker.Click(second.Center(), screen)

	if picker.chosen != 22 {
		t.Errorf("chosen = %d, want the second item", picker.chosen)
	}

	picker.chosen = 0
	picker.Click(gpu.At(-1, -1), screen)

	if picker.chosen != 0 {
		t.Errorf("chosen = %d, want nothing for a click outside every cell", picker.chosen)
	}
}
