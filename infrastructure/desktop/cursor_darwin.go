package desktop

/*
#cgo LDFLAGS: -framework CoreGraphics
void ribbon_cursor(int *x, int *y);
void test_warp(int x, int y);
*/
import "C"

import (
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/cocoamain"
)

// Cursor answers where the pointer is in points, counted from the menu-bar display's top-left corner
// as the ribbon is placed; also whether it could be read, which on macOS it always can. The corner
// grip's drag reads it rather than the page's pointer events, which were measured on Windows to jump
// backwards while the window was resized under them (FR-623). It is read on AppKit's main thread.
func Cursor() (placement.Point, bool) {
	var x, y C.int
	cocoamain.Do(func() { C.ribbon_cursor(&x, &y) })
	return placement.Point{X: int(x), Y: int(y)}, true
}

// warpPointer moves the pointer to at, in the same points Cursor answers; for the tests alone.
func warpPointer(at placement.Point) {
	cocoamain.Do(func() { C.test_warp(C.int(at.X), C.int(at.Y)) })
}
