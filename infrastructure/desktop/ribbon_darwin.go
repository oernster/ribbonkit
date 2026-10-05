package desktop

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include <stdint.h>
void *ribbon_find(const char *title);
void ribbon_place(void *ribbon, int x, int y, int width, int height);
void ribbon_frame(void *ribbon, int *x, int *y, int *width, int *height);
int ribbon_pointer_inside(void *ribbon);
void *test_window(const char *title);
void test_window_close(void *ribbon);
*/
import "C"

import (
	"io"
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/cocoamain"
)

// dragThreshold is how far the pointer must move, in DIP, before a press becomes a drag. macOS
// publishes no threshold of its own, so this is the distance Windows uses at its base DPI, which
// makes a press become a drag at the same distance on both.
var dragThreshold = placement.Size{Width: 4, Height: 4}

// FindRibbon answers the ribbon's window. On macOS there is no window class to find it by, so
// class is not used: the ribbon is the process's window titled name, the product's name, as Wails
// creates it.
func FindRibbon(_, name string) (Window, error) {
	title := C.CString(name)
	defer C.free(unsafe.Pointer(title))
	return findWith(name, func() unsafe.Pointer {
		var found unsafe.Pointer
		cocoamain.Do(func() { found = C.ribbon_find(title) })
		return found
	})
}

// HideFromTaskbar does nothing on macOS, where a ribbon keeps its Dock icon (FR-101, Amendment
// 36): Wails makes the application a regular one as it launches; making it an accessory after that
// never took the icon away on a real Mac.
func HideFromTaskbar(Window) error { return nil }

// KeepOnDisplays does nothing on macOS, where AppKit carries a drag through and offers no say in
// it while it lasts. A ribbon dragged partly off every display is put back when the move ends,
// through the same path as on Windows (FR-404, FR-406).
func KeepOnDisplays(Window, io.Writer) error { return nil }

// Place moves and sizes the ribbon in points, in one step (FR-405).
func Place(ribbon Window, at placement.Point, size placement.Size) error {
	window, err := pointerOf(ribbon)
	if err != nil {
		return err
	}
	notePlaced(ribbon, at)
	cocoamain.Do(func() {
		C.ribbon_place(window, C.int(at.X), C.int(at.Y), C.int(size.Width), C.int(size.Height))
	})
	return nil
}

// Position answers the ribbon's top-left corner in points.
func Position(ribbon Window) (placement.Point, error) {
	at, _, err := frameOf(ribbon)
	return at, err
}

// pointerInside answers whether the pointer is on the ribbon's window now (FR-615, FR-616), read on
// AppKit's main thread, which owns the window's frame.
func pointerInside(ribbon Window) (bool, error) {
	var inside bool
	err := onWindow(ribbon, func(window unsafe.Pointer) { inside = C.ribbon_pointer_inside(window) != 0 })
	return inside, err
}

// DragThreshold answers how far the pointer must move before a press becomes a drag (FR-401).
func DragThreshold() placement.Size { return dragThreshold }

// frameOf answers where the ribbon's top-left corner stands and its size, in points.
func frameOf(ribbon Window) (placement.Point, placement.Size, error) {
	var x, y, width, height C.int
	err := onWindow(ribbon, func(window unsafe.Pointer) { C.ribbon_frame(window, &x, &y, &width, &height) })
	return placement.Point{X: int(x), Y: int(y)}, placement.Size{Width: int(width), Height: int(height)}, err
}

// onWindow runs act on the NSWindow ribbon stands for, on AppKit's main thread.
func onWindow(ribbon Window, act func(unsafe.Pointer)) error {
	window, err := pointerOf(ribbon)
	if err != nil {
		return err
	}
	cocoamain.Do(func() { act(window) })
	return nil
}

// size answers the ribbon's size in points.
func size(ribbon Window) (placement.Size, error) {
	_, got, err := frameOf(ribbon)
	return got, err
}

// newTestWindow shows a window set up as Wails sets up the ribbon's, for the tests: nothing in the
// application makes one.
func newTestWindow(name string) Window {
	title := C.CString(name)
	defer C.free(unsafe.Pointer(title))
	var window unsafe.Pointer
	cocoamain.Do(func() { window = C.test_window(title) })
	return remember(window)
}

// closeTestWindow closes a window newTestWindow made.
func closeTestWindow(ribbon Window) {
	_ = onWindow(ribbon, func(window unsafe.Pointer) { C.test_window_close(window) })
	forget(ribbon)
}
