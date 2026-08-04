package gpu

import "github.com/go-gl/gl/v3.3-core/gl"

// Surface is a texture whose pixels arrive every frame rather than once at load,
// which is what a rendered page needs. It takes BGRA rows straight from the
// engine, so nothing is converted on the way to the card.
type Surface struct {
	texture Texture
	width   int
	height  int
}

func NewSurface(width, height int) *Surface {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(width), int32(height), 0, gl.BGRA, gl.UNSIGNED_BYTE, nil)

	return &Surface{
		texture: Texture{id: id},
		width:   width,
		height:  height,
	}
}

func (s *Surface) Size() (width, height int) {
	return s.width, s.height
}

// Resize throws the old pixels away, so the caller has to repaint the whole
// page rather than trusting the region the engine reports as changed
func (s *Surface) Resize(width, height int) {
	if width == s.width && height == s.height {
		return
	}

	gl.BindTexture(gl.TEXTURE_2D, s.texture.id)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(width), int32(height), 0, gl.BGRA, gl.UNSIGNED_BYTE, nil)

	s.width, s.height = width, height
}

// Upload copies one rectangle of a larger buffer, so a frame that changed a
// single row of a menu costs a single row of bandwidth.
//
// The unpack state is global to the context and is reset on the way out; left
// set, it silently corrupts every other texture load in the frame.
func (s *Surface) Upload(pixels []byte, stride, left, top, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	gl.BindTexture(gl.TEXTURE_2D, s.texture.id)

	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, int32(stride/4))
	gl.PixelStorei(gl.UNPACK_SKIP_PIXELS, int32(left))
	gl.PixelStorei(gl.UNPACK_SKIP_ROWS, int32(top))

	gl.TexSubImage2D(gl.TEXTURE_2D, 0,
		int32(left), int32(top), int32(width), int32(height),
		gl.BGRA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.PixelStorei(gl.UNPACK_SKIP_PIXELS, 0)
	gl.PixelStorei(gl.UNPACK_SKIP_ROWS, 0)
}

func (s *Surface) Texture() *Texture {
	return &s.texture
}

func (s *Surface) Delete() {
	s.texture.Delete()
}
