package gpu

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Window struct {
	handle   *glfw.Window
	width    int
	height   int
	cursorX  float64
	cursorY  float64
	scrolled float64
	typed    []rune
}

func OpenWindow(title string, width, height int, visible bool) (*Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("gpu: glfw init: %w", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	if !visible {
		glfw.WindowHint(glfw.Visible, glfw.False)
	}

	handle, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		glfw.Terminate()

		return nil, fmt.Errorf("gpu: create window: %w", err)
	}
	handle.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		glfw.Terminate()

		return nil, fmt.Errorf("gpu: gl init: %w", err)
	}
	gl.Enable(gl.DEPTH_TEST)

	window := &Window{handle: handle, width: width, height: height}
	handle.SetScrollCallback(func(_ *glfw.Window, _, dy float64) {
		window.scrolled += dy
	})
	handle.SetCharCallback(func(_ *glfw.Window, char rune) {
		window.typed = append(window.typed, char)
	})

	// GL does not follow the window on its own: without this the viewport keeps
	// the size it was created with and everything drawn after a resize is skewed
	handle.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		window.resize(width, height)
	})
	window.resize(handle.GetFramebufferSize())

	return window, nil
}

func (w *Window) resize(width, height int) {
	w.width, w.height = width, height
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (w *Window) ShouldClose() bool {
	return w.handle.ShouldClose()
}

func (w *Window) Clear(r, g, b float32) {
	gl.ClearColor(r, g, b, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (w *Window) ClearDepth() {
	gl.Clear(gl.DEPTH_BUFFER_BIT)
}

func (w *Window) Overlay(enabled bool) {
	if enabled {
		gl.Disable(gl.DEPTH_TEST)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

		return
	}

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
}

func (w *Window) Present() {
	w.handle.SwapBuffers()
	glfw.PollEvents()
}

func (w *Window) Viewport() Rect {
	return RectAt(Point{}, float32(w.width), float32(w.height))
}

func (w *Window) Close() {
	glfw.Terminate()
}

func (w *Window) Capture(path string) error {
	gl.Finish()

	pixels := make([]byte, w.width*w.height*4)
	gl.ReadPixels(0, 0, int32(w.width), int32(w.height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	img := image.NewRGBA(image.Rect(0, 0, w.width, w.height))
	stride := w.width * 4
	for y := range w.height {
		copy(img.Pix[y*stride:(y+1)*stride], pixels[(w.height-1-y)*stride:])
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	if err := png.Encode(file, img); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}
