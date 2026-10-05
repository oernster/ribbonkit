package desktop

// The icon in the macOS menu bar, the platform's notification area (FR-501 to FR-503). A click
// opens its menu, as every icon there does; the menu is rebuilt from the desktop's menu function
// each time it opens.

/*
#include <stdint.h>
#include <stdlib.h>
void tray_start(uintptr_t handle, const void *png, int length, const char *tooltip);
void tray_stop(void);
int tray_shown(void);
*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"

	"github.com/oernster/ribbonkit/infrastructure/cocoamain"
)

// errNoIcon is answered when the desktop was given no image for the icon.
var errNoIcon = errors.New("no image was given for the menu bar icon")

// tray is the ribbon's icon in the menu bar, holding the desktop its menu is built from.
type tray struct{ handle cgo.Handle }

// startTray puts the icon in the menu bar once AppKit's loop runs, answering why it could not.
func startTray(d *Desktop, icon []byte) (*tray, error) {
	if len(icon) == 0 {
		return nil, errNoIcon
	}
	t := &tray{handle: cgo.NewHandle(d)}
	tooltip := C.CString(d.app.Name)
	defer C.free(unsafe.Pointer(tooltip))
	C.tray_start(C.uintptr_t(t.handle), unsafe.Pointer(&icon[0]), C.int(len(icon)), tooltip)
	return t, nil
}

// stop takes the icon out of the menu bar.
func (t *tray) stop() {
	cocoamain.Do(func() { C.tray_stop() })
	t.handle.Delete()
}

// shownInMenuBar answers whether the icon stands in the menu bar with its image.
func shownInMenuBar() bool {
	var shown bool
	cocoamain.Do(func() { shown = C.tray_shown() != 0 })
	return shown
}

//export trayMenuNeedsUpdate
func trayMenuNeedsUpdate(handle C.uintptr_t, menu unsafe.Pointer) {
	d := cgo.Handle(handle).Value().(*Desktop)
	items := d.menu()
	showing(d, items)
	next := 0
	build(menu, items, &next)
}
