package ultralight

/*
#cgo CFLAGS: -I${SRCDIR}/sdk/include
#cgo windows LDFLAGS: -L${SRCDIR}/sdk/windows/lib -lUltralight -lUltralightCore -lAppCore -lWebCore
#cgo linux LDFLAGS: -L${SRCDIR}/sdk/linux/bin -lUltralight -lUltralightCore -lAppCore -lWebCore -Wl,-rpath,${SRCDIR}/sdk/linux/bin

#include <stdlib.h>
#include <Ultralight/CAPI.h>
#include <AppCore/CAPI.h>
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime/cgo"
	"sync"
	"unsafe"
)

var errNoRenderer = errors.New("ultralight: the renderer refused to start")

type Renderer struct {
	handle    C.ULRenderer
	resources string
}

var (
	start  sync.Once
	shared *Renderer
	failed error
)

// Open starts the engine against a resources directory, the one holding
// icudt67l.dat and cacert.pem.
//
// The engine is one per process. WebKit keeps global state that survives a
// destroyed renderer, and building a second one after any view has existed
// segfaults, so every call hands back the same renderer and there is no way to
// shut it down short of exiting. It also belongs to the thread that opened it,
// the same rule the window already follows.
//
// The path is split internally on purpose: Ultralight wants a file system root
// and a prefix within it, and handing it the resources directory as the root
// makes it look for resources/resources. That mistake aborts the process with
// no error and nothing on stdout, so the API takes the one path a caller knows.
func Open(resources string) (*Renderer, error) {
	cleaned := filepath.Clean(resources)

	start.Do(func() {
		shared, failed = open(cleaned)
	})

	if failed != nil {
		return nil, failed
	}
	if shared.resources != cleaned {
		return nil, fmt.Errorf("ultralight: the engine is already open on %s and cannot move to %s", shared.resources, cleaned)
	}

	return shared, nil
}

func open(resources string) (*Renderer, error) {
	root, prefix := filepath.Split(resources)
	if root == "" {
		root = "."
	}

	base := text(root)
	defer C.ulDestroyString(base)

	C.ulEnablePlatformFileSystem(base)
	C.ulEnablePlatformFontLoader()

	config := C.ulCreateConfig()
	defer C.ulDestroyConfig(config)

	within := text(prefix + "/")
	defer C.ulDestroyString(within)

	C.ulConfigSetResourcePathPrefix(config, within)

	C.ul_survive_thread_naming()

	handle := C.ulCreateRenderer(config)
	if handle == nil {
		return nil, errNoRenderer
	}

	opened := Renderer{
		handle:    handle,
		resources: resources,
	}

	return &opened, nil
}

// Log sends the engine's own diagnostics to a file. Worth turning on while
// integrating: most failures are reported there and nowhere else.
func Log(path string) {
	written := text(path)
	defer C.ulDestroyString(written)

	C.ulEnableDefaultLogger(written)
}

// Update lets timers, scripts and animations advance; Render paints whatever
// changed. Both belong on the thread that made the renderer.
func (r *Renderer) Update() {
	C.ulUpdate(r.handle)
}

func (r *Renderer) Render() {
	C.ulRender(r.handle)
}

type View struct {
	handle C.ULView
	width  int
	height int
	sends  func(name string, args []string)
	owner  cgo.Handle
}

func (r *Renderer) View(width, height int) *View {
	settings := C.ulCreateViewConfig()
	defer C.ulDestroyViewConfig(settings)

	// the bitmap surface is what lets the pixels be read back and uploaded as a
	// texture; an accelerated view hands back no surface at all
	C.ulViewConfigSetIsAccelerated(settings, C.bool(false))
	C.ulViewConfigSetIsTransparent(settings, C.bool(true))

	return &View{
		handle: C.ulCreateView(r.handle, C.uint(width), C.uint(height), settings, nil),
		width:  width,
		height: height,
	}
}

func (v *View) Close() {
	if v.handle == nil {
		return
	}

	if v.owner != 0 {
		v.owner.Delete()
		v.owner = 0
	}

	C.ulDestroyView(v.handle)
	v.handle = nil
}

func (v *View) LoadHTML(html string) {
	written := text(html)
	defer C.ulDestroyString(written)

	C.ulViewLoadHTML(v.handle, written)
}

func (v *View) LoadURL(url string) {
	written := text(url)
	defer C.ulDestroyString(written)

	C.ulViewLoadURL(v.handle, written)
}

func (v *View) Resize(width, height int) {
	if width == v.width && height == v.height {
		return
	}

	C.ulViewResize(v.handle, C.uint(width), C.uint(height))
	v.width, v.height = width, height
}

func (v *View) Size() (width, height int) {
	return v.width, v.height
}

// Loading reports whether the page is still coming up, which is how a caller
// knows to keep pumping Update before trusting the first frame
func (v *View) Loading() bool {
	return bool(C.ulViewIsLoading(v.handle))
}

// Rect is the region of the view that changed since it was last cleared
type Rect struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func (r Rect) Empty() bool {
	return r.Right <= r.Left || r.Bottom <= r.Top
}

func (r Rect) Width() int {
	return r.Right - r.Left
}

func (r Rect) Height() int {
	return r.Bottom - r.Top
}

// Dirty answers what changed, so a frame uploads the region that moved rather
// than the whole page
func (v *View) Dirty() Rect {
	surface := C.ulViewGetSurface(v.handle)
	if surface == nil {
		return Rect{}
	}

	bounds := C.ulSurfaceGetDirtyBounds(surface)

	return Rect{
		Left:   int(bounds.left),
		Top:    int(bounds.top),
		Right:  int(bounds.right),
		Bottom: int(bounds.bottom),
	}
}

func (v *View) Clean() {
	surface := C.ulViewGetSurface(v.handle)
	if surface == nil {
		return
	}

	C.ulSurfaceClearDirtyBounds(surface)
}

// Pixels hands the painted page to read, as BGRA rows of stride bytes. The
// buffer is only valid for the duration of the call: it is the engine's own
// memory, locked while inside and released on return.
func (v *View) Pixels(read func(pixels []byte, stride int)) {
	surface := C.ulViewGetSurface(v.handle)
	if surface == nil {
		return
	}

	bitmap := C.ulBitmapSurfaceGetBitmap(C.ULBitmapSurface(surface))
	if bitmap == nil || bool(C.ulBitmapIsEmpty(bitmap)) {
		return
	}

	raw := C.ulBitmapLockPixels(bitmap)
	defer C.ulBitmapUnlockPixels(bitmap)

	stride := int(C.ulBitmapGetRowBytes(bitmap))
	size := int(C.ulBitmapGetSize(bitmap))

	read(unsafe.Slice((*byte)(raw), size), stride)
}

// WritePNG saves the current frame, which is the cheapest way to check a layout
// without a window
func (v *View) WritePNG(path string) bool {
	surface := C.ulViewGetSurface(v.handle)
	if surface == nil {
		return false
	}

	bitmap := C.ulBitmapSurfaceGetBitmap(C.ULBitmapSurface(surface))
	if bitmap == nil {
		return false
	}

	written := C.CString(path)
	defer C.free(unsafe.Pointer(written))

	return bool(C.ulBitmapWritePNG(bitmap, written))
}

func text(value string) C.ULString {
	written := C.CString(value)
	defer C.free(unsafe.Pointer(written))

	return C.ulCreateString(written)
}

func Version() string {
	return C.GoString(C.ulVersionString())
}
