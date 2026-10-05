package desktop

// Linux is told of the pointer crossing the ribbon's window by GTK rather than reading where it is:
// under XWayland an X11 window cannot see the pointer leave for a Wayland one, neither in the page
// nor by asking the X server, while GTK's crossing events report it (measured 2026-09-28,
// REQUIREMENTS section 2.3). desktop_linux.c connects them; desktopPointer receives them.

import "C"

import (
	"runtime/cgo"
	"sync"
)

// pointerTracker is whether crossings are let through.
type pointerTracker struct {
	guard sync.Mutex
	on    bool
}

// TrackPointer starts or stops reporting the pointer arriving on the ribbon and leaving it (FR-615,
// FR-616). The crossings are always connected; this only lets them through.
func (d *Desktop) TrackPointer(_ Window, on bool) {
	d.pointer.guard.Lock()
	defer d.pointer.guard.Unlock()
	d.pointer.on = on
}

//export desktopPointer
func desktopPointer(handle uintptr, inside C.int) {
	d := cgo.Handle(handle).Value().(*Desktop)
	d.pointer.guard.Lock()
	on := d.pointer.on
	d.pointer.guard.Unlock()
	if on {
		d.send(pointerEvent(inside != 0))
	}
}
