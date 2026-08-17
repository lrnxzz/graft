package gpu

import "github.com/go-gl/glfw/v3.3/glfw"

type Key int

const (
	KeyW         Key = Key(glfw.KeyW)
	KeyA         Key = Key(glfw.KeyA)
	KeyS         Key = Key(glfw.KeyS)
	KeyD         Key = Key(glfw.KeyD)
	KeyE         Key = Key(glfw.KeyE)
	KeyF         Key = Key(glfw.KeyF)
	KeyT         Key = Key(glfw.KeyT)
	KeyP         Key = Key(glfw.KeyP)
	KeyV         Key = Key(glfw.KeyV)
	KeySlash     Key = Key(glfw.KeySlash)
	KeyPeriod    Key = Key(glfw.KeyPeriod)
	KeyEnter     Key = Key(glfw.KeyEnter)
	KeyEscape    Key = Key(glfw.KeyEscape)
	KeyBackspace Key = Key(glfw.KeyBackspace)
	KeyTab       Key = Key(glfw.KeyTab)
	KeyUp        Key = Key(glfw.KeyUp)
	KeyDown      Key = Key(glfw.KeyDown)
	KeySpace     Key = Key(glfw.KeySpace)
	KeyShift     Key = Key(glfw.KeyLeftShift)
	KeyCtrl      Key = Key(glfw.KeyLeftControl)
	KeyAlt       Key = Key(glfw.KeyLeftAlt)
)

func Digit(index int) Key {
	return Key(glfw.Key1 + glfw.Key(index))
}

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Letter(name rune) Key {
	return Key(glfw.KeyA + glfw.Key(name-'A'))
}

var named = map[Key]string{
	KeySlash:     "/",
	KeyPeriod:    ".",
	KeyEnter:     "Enter",
	KeyEscape:    "Escape",
	KeyBackspace: "Backspace",
	KeyTab:       "Tab",
	KeyUp:        "Up",
	KeyDown:      "Down",
	KeySpace:     "Space",
	KeyShift:     "Shift",
	KeyCtrl:      "Ctrl",
	KeyAlt:       "Alt",
}

// Keys is every key the viewer knows how to name, which is what a screen sees
// forwarded to it and what a plugin may claim
var Keys = keyboard()

var keyNames = naming()

func keyboard() []Key {
	keys := make([]Key, 0, len(letters)+len(named))

	for _, letter := range letters {
		keys = append(keys, Letter(letter))
	}
	for key := range named {
		keys = append(keys, key)
	}

	return keys
}

func naming() map[Key]string {
	names := make(map[Key]string, len(letters)+len(named))

	for _, letter := range letters {
		names[Letter(letter)] = string(letter)
	}
	for key, name := range named {
		names[key] = name
	}

	return names
}

func Name(key Key) string {
	return keyNames[key]
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

// Clipboard is what a paste would insert, which the chat needs because a
// keyboard without a slash cannot otherwise write a server command
func (w *Window) Clipboard() string {
	return w.handle.GetClipboardString()
}
