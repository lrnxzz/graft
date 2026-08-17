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

// Open takes the directory holding icudt67l.dat and cacert.pem, and splits it
// into the root and prefix the engine wants: passing it whole aborts the process
// with nothing on stdout. One renderer per process, on the thread that opened it.
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

// most of the engine's failures are reported here and nowhere else
func Log(path string) {
	written := text(path)
	defer C.ulDestroyString(written)

	C.ulEnableDefaultLogger(written)
}

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
	if v.handle != nil {
		C.ulDestroyView(v.handle)
		v.handle = nil
	}

	// the handle goes only after the view: a callback firing during the destroy
	// still resolves it, and a view that never created still lets it go
	if v.owner != 0 {
		v.owner.Delete()
		v.owner = 0
	}
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

func (v *View) Loading() bool {
	return bool(C.ulViewIsLoading(v.handle))
}

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

// Pixels reads BGRA rows of stride bytes. The buffer is the engine's own, locked
// for the call and released on return, so it must not outlive it.
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
