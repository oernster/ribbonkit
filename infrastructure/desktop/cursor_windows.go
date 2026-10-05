package desktop

import (
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
)

// Cursor answers where the pointer is on the virtual desktop in physical pixels; also whether
// it could be read. The corner grip's drag reads it rather than the page's pointer events, whose
// screen position was measured to jump backwards while the window was resized under it (FR-623,
// 2026-10-04: the page sent 89, 85, 90 while the cursor only moved one way).
func Cursor() (placement.Point, bool) {
	var cursor point
	if ok, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor))); ok == 0 {
		return placement.Point{}, false
	}
	return placement.Point{X: int(cursor.x), Y: int(cursor.y)}, true
}
