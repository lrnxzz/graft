#include "bridge.h"

#include <stdlib.h>
#include <string.h>
#include <JavaScriptCore/JSObjectRef.h>
#include <JavaScriptCore/JSStringRef.h>
#include <JavaScriptCore/JSValueRef.h>

// implemented on the Go side of the package
extern void goSend(uintptr_t owner, char **words, int count);

// text copies a JavaScript value out as a C string the caller has to free
static char *text(JSContextRef ctx, JSValueRef value) {
	JSStringRef written = JSValueToStringCopy(ctx, value, NULL);
	if (written == NULL) {
		return NULL;
	}

	size_t room = JSStringGetMaximumUTF8CStringSize(written);
	char *copied = malloc(room);
	if (copied != NULL) {
		JSStringGetUTF8CString(written, copied, room);
	}

	JSStringRelease(written);

	return copied;
}

// dispatch is what the page reaches when it calls gocraft.send(name, ...). Every
// argument is handed over as text, which is all a DOM attribute can carry.
static JSValueRef dispatch(JSContextRef ctx, JSObjectRef function, JSObjectRef self,
                           size_t count, const JSValueRef arguments[], JSValueRef *exception) {
	uintptr_t owner = (uintptr_t)JSObjectGetPrivate(function);
	if (owner == 0 || count == 0) {
		return JSValueMakeUndefined(ctx);
	}

	char **words = calloc(count, sizeof(char *));
	if (words == NULL) {
		return JSValueMakeUndefined(ctx);
	}

	for (size_t index = 0; index < count; index++) {
		words[index] = text(ctx, arguments[index]);
	}

	goSend(owner, words, (int)count);

	for (size_t index = 0; index < count; index++) {
		free(words[index]);
	}
	free(words);

	return JSValueMakeUndefined(ctx);
}

// the class exists only so the function can carry its owner as private data;
// JSObjectMakeFunctionWithCallback has nowhere to put it
static JSClassRef sending(void) {
	static JSClassRef shared = NULL;
	if (shared != NULL) {
		return shared;
	}

	JSClassDefinition described = kJSClassDefinitionEmpty;
	described.className = "gocraftSend";
	described.callAsFunction = dispatch;

	shared = JSClassCreate(&described);

	return shared;
}

static void install(void *user_data, ULView caller, unsigned long long frame_id,
                    bool is_main_frame, ULString url) {
	if (!is_main_frame) {
		return;
	}

	JSContextRef ctx = ulViewLockJSContext(caller);
	JSObjectRef window = JSContextGetGlobalObject(ctx);

	JSObjectRef bridge = JSObjectMake(ctx, NULL, NULL);
	JSObjectRef sender = JSObjectMake(ctx, sending(), (void *)(uintptr_t)user_data);

	JSStringRef name = JSStringCreateWithUTF8CString("send");
	JSObjectSetProperty(ctx, bridge, name, sender, kJSPropertyAttributeReadOnly, NULL);
	JSStringRelease(name);

	JSStringRef global = JSStringCreateWithUTF8CString("gocraft");
	JSObjectSetProperty(ctx, window, global, bridge, kJSPropertyAttributeReadOnly, NULL);
	JSStringRelease(global);

	ulViewUnlockJSContext(caller);
}

void ul_listen(ULView view, uintptr_t owner) {
	ulViewSetWindowObjectReadyCallback(view, install, (void *)owner);
}

#ifdef _WIN32
#include <windows.h>

// The engine is built with MSVC, which names a thread by raising an exception
// that only a debugger is meant to catch. Nothing catches it here, so it reaches
// the Go runtime's last-resort handler, which prints a stack trace and kills the
// process the moment the renderer spawns its threads. It is continuable by
// design: whoever raised it expects to carry on.
#define THREAD_NAMING 0x406D1388

static volatile LONG seen = 0;
static volatile LONG caught = 0;

static LONG CALLBACK carry_on(EXCEPTION_POINTERS *reported) {
	InterlockedIncrement(&seen);

	if (reported->ExceptionRecord->ExceptionCode == THREAD_NAMING) {
		InterlockedIncrement(&caught);

		return EXCEPTION_CONTINUE_EXECUTION;
	}

	return EXCEPTION_CONTINUE_SEARCH;
}

static PVOID guard = NULL;
static PVOID fallback = NULL;

// Both are registered because only the second one works. Counting the calls
// showed the exception handler is entered and its answer to continue is ignored,
// so what actually saves the process is the continue handler, which still runs
// ahead of the runtime's own. The first-chance registration is kept as the line
// that is supposed to be enough, and -check reports if that ever changes.
void ul_survive_thread_naming(void) {
	if (guard == NULL) {
		guard = AddVectoredExceptionHandler(1, carry_on);
	}
	if (fallback == NULL) {
		fallback = AddVectoredContinueHandler(1, carry_on);
	}
}

int ul_guard_installed(void) {
	return guard != NULL && fallback != NULL;
}

int ul_exceptions_seen(void) {
	return seen;
}

int ul_exceptions_caught(void) {
	return caught;
}

void ul_name_a_thread(void) {
	ULONG_PTR described[4] = {0x1000, (ULONG_PTR) "gocraft-probe", (ULONG_PTR)-1, 0};

	RaiseException(THREAD_NAMING, 0, 4, described);
}
#else
void ul_survive_thread_naming(void) {}

int ul_guard_installed(void) {
	return 0;
}

void ul_name_a_thread(void) {}

int ul_exceptions_seen(void) {
	return 0;
}

int ul_exceptions_caught(void) {
	return 0;
}
#endif
