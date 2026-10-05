package desktop

import (
	"testing"
	"unsafe"
)

// FR-623: the pointer is read where Windows says it is.
func TestTheCursorIsWhereWindowsSaysItIs(t *testing.T) {
	var want point
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&want)))
	got, ok := Cursor()
	if !ok {
		t.Fatal("the cursor could not be read")
	}
	if got.X != int(want.x) || got.Y != int(want.y) {
		t.Errorf("read %+v, Windows says %d,%d", got, want.x, want.y)
	}
}
