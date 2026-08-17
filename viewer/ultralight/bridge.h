#ifndef GRAFT_ULTRALIGHT_BRIDGE_H
#define GRAFT_ULTRALIGHT_BRIDGE_H

#include <stdint.h>
#include <Ultralight/CAPI.h>

// ul_listen makes every page this view loads carry a graft.send function that
// calls back into Go. The owner is a cgo handle, and it rides on the function
// object itself rather than on the page, so a stale reference cannot make one
// view's message arrive as another's.
void ul_listen(ULView view, uintptr_t owner);

// ul_survive_thread_naming keeps the process alive when the engine names its own
// threads. It has to be called before the renderer starts.
void ul_survive_thread_naming(void);

// ul_guard_installed reports whether the handler took, and ul_name_a_thread
// raises the exception on purpose so the guard can be proven before the engine
// is involved. Returning at all means it held.
int ul_guard_installed(void);
void ul_name_a_thread(void);

int ul_exceptions_seen(void);
int ul_exceptions_caught(void);

#endif
