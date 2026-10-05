// Package gtkmain runs work on GTK's main loop, the one thread GTK may be called from. Wails runs
// that loop on the process's main thread while the facade's calls arrive on other goroutines, so
// every GTK call the desktop makes goes through Do.
package gtkmain

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
void gtkmain_invoke(guintptr handle);
gboolean gtkmain_init(void);
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/cgo"
)

// errNoDisplay is answered when GTK cannot open a display, as in a session with no desktop.
var errNoDisplay = errors.New("GTK could not open a display")

// The main goroutine is locked to the process's first thread before main or a test starts. The
// loop Serve runs then sits on the thread GTK was opened on, as the one Wails runs does.
func init() { runtime.LockOSThread() }

// Do runs f on GTK's main loop and waits for it to return. Called on the loop's own thread it runs
// f at once; from anywhere else f waits its turn in the loop. A panic in f is raised again here, on
// the caller's goroutine, where the caller's own recover can see it.
func Do(f func()) {
	done := make(chan any, 1)
	handle := cgo.NewHandle(func() {
		defer func() { done <- recover() }()
		f()
	})
	C.gtkmain_invoke(C.guintptr(handle))
	if failure := <-done; failure != nil {
		panic(fmt.Sprintf("on GTK's main loop: %v", failure))
	}
}

//export gtkmainRun
func gtkmainRun(handle uintptr) {
	held := cgo.Handle(handle)
	f := held.Value().(func())
	held.Delete()
	f()
}

// Serve opens GTK, runs body on another goroutine while the loop runs on this one, then answers
// what body answered. It is for a process that owns the loop itself, as a test binary does; the
// application's loop is Wails'. It must be called from the main goroutine.
func Serve(body func() int) (int, error) {
	if C.gtkmain_init() == 0 {
		return 0, errNoDisplay
	}
	result := make(chan int, 1)
	go func() {
		code := body()
		result <- code
		Do(func() { C.gtk_main_quit() })
	}()
	C.gtk_main()
	return <-result, nil
}

// ServeTests runs a test binary's tests on X11 with GTK's loop running, answering the exit code.
// Without a display there is nothing the tests could measure, so that is a failure, said plainly.
func ServeTests(run func() int) int {
	ForceX11()
	code, err := Serve(run)
	if err != nil {
		fmt.Fprintf(os.Stderr, "these tests need a desktop session: %v\n", err)
		return 1
	}
	return code
}
