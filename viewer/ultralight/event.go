package ultralight

/*
#include <Ultralight/CAPI.h>
*/
import "C"

type Button int

const (
	NoButton Button = iota
	Left
	Middle
	Right
)

// Modifier is the set of qualifier keys held while another key is struck. The C
// header takes them as a bare unsigned int; these are the bits the engine reads.
type Modifier uint

const (
	Shift Modifier = 1 << iota
	Ctrl
	Alt
	Meta
)

func (v *View) Moved(x, y int) {
	v.mouse(C.kMouseEventType_MouseMoved, x, y, NoButton)
}

func (v *View) Pressed(x, y int, button Button) {
	v.mouse(C.kMouseEventType_MouseDown, x, y, button)
}

func (v *View) Released(x, y int, button Button) {
	v.mouse(C.kMouseEventType_MouseUp, x, y, button)
}

func (v *View) mouse(kind C.ULMouseEventType, x, y int, button Button) {
	event := C.ulCreateMouseEvent(kind, C.int(x), C.int(y), C.ULMouseButton(button))
	defer C.ulDestroyMouseEvent(event)

	C.ulViewFireMouseEvent(v.handle, event)
}

// Scrolled moves the page by pixels, the same unit the wheel already reports
func (v *View) Scrolled(dx, dy int) {
	event := C.ulCreateScrollEvent(C.kScrollEventType_ScrollByPixel, C.int(dx), C.int(dy))
	defer C.ulDestroyScrollEvent(event)

	C.ulViewFireScrollEvent(v.handle, event)
}

// Typed delivers text the window already composed, which is the only way accents
// and layouts reach an input; a key code alone would spell them wrong
func (v *View) Typed(typed string, held Modifier) {
	written := text(typed)
	defer C.ulDestroyString(written)

	v.key(C.kKeyEventType_Char, 0, held, written)
}

func (v *View) KeyDown(code int, held Modifier) {
	empty := text("")
	defer C.ulDestroyString(empty)

	v.key(C.kKeyEventType_RawKeyDown, code, held, empty)
}

func (v *View) KeyUp(code int, held Modifier) {
	empty := text("")
	defer C.ulDestroyString(empty)

	v.key(C.kKeyEventType_KeyUp, code, held, empty)
}

func (v *View) key(kind C.ULKeyEventType, code int, held Modifier, written C.ULString) {
	event := C.ulCreateKeyEvent(kind, C.uint(held), C.int(code), 0,
		written, written, C.bool(false), C.bool(false), C.bool(false))
	defer C.ulDestroyKeyEvent(event)

	C.ulViewFireKeyEvent(v.handle, event)
}

// The engine reads Windows virtual key codes even where it is not running on
// Windows. Letters, digits and space already agree with GLFW's own numbering,
// which is ASCII for all three; the rest have to be named.
var virtualKeys = map[int]int{
	257: 0x0D, // enter
	256: 0x1B, // escape
	259: 0x08, // backspace
	258: 0x09, // tab
	340: 0x10, // shift
	341: 0x11, // ctrl
	47:  0xBF, // slash
}

// Virtual translates a window's key code into the one the engine reads. It takes
// a plain int rather than a key type so the engine keeps knowing nothing about
// the window that feeds it.
func Virtual(key int) int {
	named, renamed := virtualKeys[key]
	if renamed {
		return named
	}

	return key
}

// Focus decides whether the page believes it has the caret, which is what makes
// an input show one and accept the keys that follow
func (v *View) Focus(focused bool) {
	if focused {
		C.ulViewFocus(v.handle)

		return
	}

	C.ulViewUnfocus(v.handle)
}
