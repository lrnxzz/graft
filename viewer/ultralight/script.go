package ultralight

/*
#include <stdlib.h>
#include <Ultralight/CAPI.h>
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// Sends has to be set before the page loads: the window object is built once per
// load, and a view nobody listens to never grows the function at all.
func (v *View) Sends(handle func(name string, args []string)) {
	v.sends = handle

	if v.owner == 0 {
		v.owner = cgo.NewHandle(v)
	}

	C.ul_listen(v.handle, C.uintptr_t(v.owner))
}

//export goSend
func goSend(owner C.uintptr_t, words **C.char, count C.int) {
	view, is := cgo.Handle(owner).Value().(*View)
	if !is || view.sends == nil || count == 0 {
		return
	}

	spoken := unsafe.Slice(words, int(count))

	said := make([]string, 0, len(spoken))
	for _, word := range spoken {
		said = append(said, C.GoString(word))
	}

	view.sends(said[0], said[1:])
}

var errScript = errors.New("ultralight: the page refused the script")

func (v *View) Eval(script string) (string, error) {
	written := text(script)
	defer C.ulDestroyString(written)

	var refused C.ULString

	// the engine always writes to the exception, leaving it empty when the script
	// ran; a nil check alone reports every successful call as a failure
	result := C.ulViewEvaluateScript(v.handle, written, &refused)
	if refused != nil && !bool(C.ulStringIsEmpty(refused)) {
		return "", errors.Join(errScript, errors.New(decode(refused)))
	}
	if result == nil {
		return "", nil
	}

	return decode(result), nil
}

func decode(value C.ULString) string {
	return C.GoStringN(C.ulStringGetData(value), C.int(C.ulStringGetLength(value)))
}

// Guard installs the handler that keeps the engine's thread naming from killing
// the process, and reports whether it took. False everywhere but Windows.
func Guard() bool {
	C.ul_survive_thread_naming()

	return C.ul_guard_installed() != 0
}

// Probe raises the exception the guard exists for; returning means it held
func Probe() {
	C.ul_name_a_thread()
}

// zero seen means the guard is registered but never consulted, which is a
// different problem from a guard that declines
func Caught() (seen, caught int) {
	return int(C.ul_exceptions_seen()), int(C.ul_exceptions_caught())
}
