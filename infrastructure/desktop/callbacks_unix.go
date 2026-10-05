//go:build linux || darwin

package desktop

// The functions the native side calls back on off Windows, from GTK's loop or AppKit's main thread.
// Each platform's C or Objective-C half includes _cgo_export.h to reach them.

import "C"

import (
	"runtime/cgo"
	"sync"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
)

// shown is the menu on screen: the desktop its choice goes to and the items it was built from. One
// menu is open at a time, as both GTK and AppKit allow.
var shown struct {
	sync.Mutex
	desktop *Desktop
	items   []menus.Item
}

// showing records that items are the menu on screen, their choice going to d.
func showing(d *Desktop, items []menus.Item) {
	shown.Lock()
	defer shown.Unlock()
	shown.desktop, shown.items = d, items
}

//export desktopMoved
func desktopMoved(handle uintptr, x, y C.int) {
	cgo.Handle(handle).Value().(*Desktop).moved(placement.Point{X: int(x), Y: int(y)})
}

//export desktopDisplaysChanged
func desktopDisplaysChanged(handle uintptr) {
	cgo.Handle(handle).Value().(*Desktop).send(Event{Kind: EventDisplayChanged})
}

//export desktopMenuClosed
func desktopMenuClosed() {
	shown.Lock()
	d := shown.desktop
	shown.Unlock()
	if d != nil {
		d.send(Event{Kind: EventMenuClosed})
	}
}

//export desktopMenuChosen
func desktopMenuChosen(index C.int) {
	shown.Lock()
	d, items := shown.desktop, shown.items
	shown.Unlock()
	if action, ok := actionAt(items, int(index)); ok && d != nil {
		d.send(Event{Kind: EventMenu, Action: action})
	}
}
