package viewer

import "github.com/lrnxzz/go-craft/viewer/gpu"

// assets are the atlases and the font. The renderer, the canvas and the hand all
// read from them, so they outlive all three and are freed here rather than by
// whichever happens to shut down last.
type assets struct {
	tileset *Tileset
	iconset *Iconset
	font    *Font
}

func loadAssets() (*assets, error) {
	tileset, err := LoadTileset()
	if err != nil {
		return nil, err
	}

	iconset, err := LoadIconset()
	if err != nil {
		return nil, err
	}

	font, err := LoadFont()
	if err != nil {
		return nil, err
	}

	loaded := assets{
		tileset: tileset,
		iconset: iconset,
		font:    font,
	}

	return &loaded, nil
}

func (a *assets) close() {
	a.font.Close()
	a.iconset.Delete()
	a.tileset.Delete()
}

// surfaces are the two ways the viewer draws: a canvas in screen space and a
// painter in world space, each with the program that puts it on the framebuffer.
type surfaces struct {
	canvas  Canvas
	painter Painter

	overlay *gpu.Program
	paint   *gpu.Program
}

func newSurfaces(loaded *assets) (surfaces, error) {
	overlay, err := gpu.NewProgram(hudVertexShader, hudFragmentShader)
	if err != nil {
		return surfaces{}, err
	}

	paint, err := gpu.NewProgram(paintVertexShader, paintFragmentShader)
	if err != nil {
		return surfaces{}, err
	}

	drawn := surfaces{
		overlay: overlay,
		paint:   paint,
	}
	drawn.canvas.font = loaded.font
	drawn.canvas.icons = loaded.iconset

	return drawn, nil
}

// world draws what sits inside the scene, with depth still on
func (s *surfaces) world(camera Camera, now float64, draw func(*Painter)) {
	s.painter.reset(camera, now)
	draw(&s.painter)
	s.painter.paint(s.paint)
}

// over draws what sits on top of it, in screen space and with depth off
func (s *surfaces) over(window *gpu.Window, now float64, draw func(*Canvas)) {
	window.Overlay(true)
	defer window.Overlay(false)

	s.canvas.reset(window.Viewport(), window.Cursor(), now)
	draw(&s.canvas)
	s.canvas.paint(s.overlay)
}

func (s *surfaces) close() {
	s.overlay.Delete()
	s.paint.Delete()
}
