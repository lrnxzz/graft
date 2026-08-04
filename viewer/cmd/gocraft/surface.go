package main

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/plugin"
	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

type surface struct {
	canvas *viewer.Canvas
	scale  float32
}

func (s surface) Size() (float32, float32) {
	screen := s.canvas.Screen()

	return screen.Width() / s.scale, screen.Height() / s.scale
}

func (s surface) Fill(x, y, width, height float32, color plugin.Color) {
	s.canvas.Fill(s.area(x, y, width, height), tint(color))
}

func (s surface) Text(text string, x, y, scale float32, color plugin.Color) {
	s.canvas.Shadowed(text, s.at(x, y), scale*s.scale, tint(color))
}

func (s surface) Icon(sprite plugin.Sprite, x, y, size float32) {
	s.canvas.Icon(item(sprite), s.area(x, y, size, size))
}

func (s surface) Measure(text string, scale float32) float32 {
	return s.canvas.TextWidth(text, scale*s.scale) / s.scale
}

func (s surface) at(x, y float32) gpu.Point {
	return gpu.At(x*s.scale, y*s.scale)
}

func (s surface) area(x, y, width, height float32) gpu.Rect {
	return gpu.RectAt(s.at(x, y), width*s.scale, height*s.scale)
}

func tint(color plugin.Color) gpu.Color {
	return gpu.RGBA(color.Red, color.Green, color.Blue, color.Alpha)
}

func item(sprite plugin.Sprite) gocraft.ItemID {
	id, known := identified(plugin.Item(sprite))
	if !known {
		return 0
	}

	return id
}
