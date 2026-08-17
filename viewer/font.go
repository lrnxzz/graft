package viewer

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"

	"github.com/lrnxzz/graft/viewer/gpu"
)

//go:embed assets/ascii.png
var fontImage []byte

const (
	glyphSize    = 8
	glyphColumns = 16
	glyphCount   = 128
	spaceWidth   = 4
)

type Font struct {
	texture *gpu.Texture
	widths  [glyphCount]float32
}

func LoadFont() (*Font, error) {
	img, err := png.Decode(bytes.NewReader(fontImage))
	if err != nil {
		return nil, err
	}

	font := &Font{texture: gpu.NewTexture(img)}
	for glyph := range font.widths {
		font.widths[glyph] = glyphWidth(img, glyph)
	}
	font.widths[' '] = spaceWidth

	return font, nil
}

func glyphWidth(img image.Image, glyph int) float32 {
	originX := glyph % glyphColumns * glyphSize
	originY := glyph / glyphColumns * glyphSize

	width := 0
	for x := range glyphSize {
		for y := range glyphSize {
			if _, _, _, alpha := img.At(originX+x, originY+y).RGBA(); alpha > 0 {
				width = x + 1

				break
			}
		}
	}
	if width == 0 {
		return 0
	}

	return float32(width + 1)
}

func (f *Font) Close() {
	f.texture.Delete()
}

func (f *Font) Bind(unit uint32) {
	f.texture.Bind(unit)
}

func (f *Font) Width(text string) float32 {
	var width float32
	for _, glyph := range text {
		if glyph < glyphCount {
			width += f.widths[glyph]
		}
	}

	return width
}

// glyph answers where a letter sits on the sheet and how far the pen advances
// past it, leaving the emitting to whoever is batching
func (f *Font) glyph(letter rune) (gpu.UV, float32, bool) {
	if letter >= glyphCount {
		return gpu.UV{}, 0, false
	}

	uv := gpu.UV{
		U0: float32(int(letter)%glyphColumns) / glyphColumns,
		V0: float32(int(letter)/glyphColumns) / glyphColumns,
	}
	uv.U1 = uv.U0 + 1.0/glyphColumns
	uv.V1 = uv.V0 + 1.0/glyphColumns

	return uv, f.widths[letter], true
}
