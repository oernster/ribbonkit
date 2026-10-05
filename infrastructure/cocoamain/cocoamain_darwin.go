// Package cocoamain runs work on AppKit's main thread, the one thread AppKit may be called from.
// Wails runs the application's loop on the process's main thread while the facade's calls arrive on
// other goroutines, so every AppKit call the desktop makes goes through Do. It is the macOS
// counterpart of gtkmain.
package cocoamain

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include <stdint.h>
void cocoamain_invoke(uintptr_t handle);
void cocoamain_serve(void);
void cocoamain_stop(void);
*/
import "C"

import (
	"fmt"
	"runtime"
	"runtime/cgo"
)

// The main goroutine is locked to the process's first thread before main or a test starts, the
// thread AppKit must run on.
func init() { runtime.LockOSThread() }

// Do runs f on AppKit's main thread and waits for it to return. Called on that thread it runs f at
// once; from anywhere else f waits its turn in the main queue, which the loop only serves once the
// application has finished launching. A panic in f is raised again here, on the caller's
// goroutine, where the caller's own recover can see it.
func Do(f func()) {
	done := make(chan any, 1)
	handle := cgo.NewHandle(func() {
		defer func() { done <- recover() }()
		f()
	})
	C.cocoamain_invoke(C.uintptr_t(handle))
	if failure := <-done; failure != nil {
		panic(fmt.Sprintf("on AppKit's main thread: %v", failure))
	}
}

//export cocoamainRun
func cocoamainRun(handle C.uintptr_t) {
	held := cgo.Handle(handle)
	f := held.Value().(func())
	held.Delete()
	f()
}

// Serve runs AppKit's loop on this thread while body runs on another goroutine, then answers what
// body answered. It is for a process that owns the loop itself, as a test binary does; the
// application's loop is Wails'. It must be called from the main goroutine.
func Serve(body func() int) int {
	result := make(chan int, 1)
	go func() {
		code := body()
		result <- code
		Do(func() { C.cocoamain_stop() })
	}()
	C.cocoamain_serve()
	return <-result
}
