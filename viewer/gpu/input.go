package gpu

import "github.com/go-gl/glfw/v3.3/glfw"

type Key int

const (
	KeyW         Key = Key(glfw.KeyW)
	KeyA         Key = Key(glfw.KeyA)
	KeyS         Key = Key(glfw.KeyS)
	KeyD         Key = Key(glfw.KeyD)
	KeyE         Key = Key(glfw.KeyE)
	KeyT         Key = Key(glfw.KeyT)
	KeyP         Key = Key(glfw.KeyP)
	KeySlash     Key = Key(glfw.KeySlash)
	KeyEnter     Key = Key(glfw.KeyEnter)
	KeyEscape    Key = Key(glfw.KeyEscape)
	KeyBackspace Key = Key(glfw.KeyBackspace)
	KeySpace     Key = Key(glfw.KeySpace)
	KeyShift     Key = Key(glfw.KeyLeftShift)
	KeyCtrl      Key = Key(glfw.KeyLeftControl)
)

func Digit(index int) Key {
	return Key(glfw.Key1 + glfw.Key(index))
}

type Button int

const (
	ButtonLeft  Button = Button(glfw.MouseButtonLeft)
	ButtonRight Button = Button(glfw.MouseButtonRight)
)

func (w *Window) Pressed(key Key) bool {
	return w.handle.GetKey(glfw.Key(key)) == glfw.Press
}

func (w *Window) Time() float64 {
	return glfw.GetTime()
}

func (w *Window) GrabCursor() {
	w.handle.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	w.cursorX, w.cursorY = w.handle.GetCursorPos()
}

func (w *Window) ReleaseCursor() {
	w.handle.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

func (w *Window) Cursor() Point {
	x, y := w.handle.GetCursorPos()

	return At(float32(x), float32(y))
}

func (w *Window) Clicking(button Button) bool {
	return w.handle.GetMouseButton(glfw.MouseButton(button)) == glfw.Press
}

func (w *Window) CursorDelta() Point {
	x, y := w.handle.GetCursorPos()
	moved := At(float32(x-w.cursorX), float32(y-w.cursorY))
	w.cursorX, w.cursorY = x, y

	return moved
}

func (w *Window) Scroll() float64 {
	scrolled := w.scrolled
	w.scrolled = 0

	return scrolled
}

func (w *Window) Typed() []rune {
	typed := w.typed
	w.typed = nil

	return typed
}
