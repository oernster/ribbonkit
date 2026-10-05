package desktop

import (
	"fmt"
	"io"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/oernster/ribbonkit/domain/placement"
)

// Win32 values the move fence uses.
const (
	wmMoving                = 0x0216 // WM_MOVING: lParam points at the proposed window rectangle
	gwlpWndProc             = -4     // GWLP_WNDPROC
	monitorDefaultToNearest = 0x2    // MONITOR_DEFAULTTONEAREST
	movingHandled           = 1      // TRUE: the rectangle was changed
)

var (
	procCallWindowProc   = user32.NewProc("CallWindowProcW")
	procMonitorFromPoint = user32.NewProc("MonitorFromPoint")
	procGetMonitorInfo   = user32.NewProc("GetMonitorInfoW")
	procRtlMoveMemory    = kernel32.NewProc("RtlMoveMemory")
)

// monitorInfo is Win32 MONITORINFO.
type monitorInfo struct {
	size    uint32
	monitor rect
	work    rect
	flags   uint32
}

// wndProcIndex is GWLP_WNDPROC held in a variable, since a negative constant cannot convert to the
// unsigned argument the call takes.
var wndProcIndex int32 = gwlpWndProc

// moveFence keeps a window wholly on one display while it is dragged: each rectangle Windows
// proposes is moved the least distance that keeps it inside the work area of the display under the
// pointer. Pushing against an edge stops there; carrying the pointer onto another display takes the
// window with it, fenced by that display instead.
type moveFence struct {
	previous uintptr
	log      io.Writer
}

// KeepOnDisplays fences the ribbon's moves (FR-401, FR-405), placing a window procedure in front of
// Wails' own. Every other message goes on to Wails unchanged.
func KeepOnDisplays(ribbon Window, log io.Writer) error {
	fence := &moveFence{log: log}
	previous, _, err := procSetWindowLongPtr.Call(uintptr(ribbon), uintptr(wndProcIndex), windows.NewCallback(fence.proc))
	if previous == 0 {
		return fmt.Errorf("fencing the ribbon's moves: %w", err)
	}
	fence.previous = previous
	return nil
}

func (f *moveFence) proc(hwnd, message, wParam, lParam uintptr) (result uintptr) {
	if message == wmMoving && f.fenceMove(lParam) {
		return movingHandled
	}
	result, _, _ = procCallWindowProc.Call(f.previous, hwnd, message, wParam, lParam)
	return result
}

// fenceMove rewrites the proposed rectangle at address; false when it could not, leaving the move to
// Wails as it was. A panic here would end the application from Windows' own thread, so it is
// recovered and logged.
func (f *moveFence) fenceMove(address uintptr) (rewritten bool) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(f.log, "fence: recovered from %v\n", failure)
			rewritten = false
		}
	}()
	var cursor point
	if ok, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor))); ok == 0 {
		return false
	}
	work, ok := workAreaAt(cursor)
	if !ok {
		return false
	}
	var proposed rect
	size := unsafe.Sizeof(proposed)
	_, _, _ = procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&proposed)), address, size)
	kept := fenced(proposed, work)
	_, _, _ = procRtlMoveMemory.Call(address, uintptr(unsafe.Pointer(&kept)), size)
	return true
}

// workAreaAt answers the work area of the display nearest the point.
func workAreaAt(at point) (placement.Rect, bool) {
	packed := uintptr(uint32(at.x)) | uintptr(uint32(at.y))<<32
	monitor, _, _ := procMonitorFromPoint.Call(packed, monitorDefaultToNearest)
	info := monitorInfo{size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, _ := procGetMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return placement.Rect{}, false
	}
	w := info.work
	return placement.Rect{Left: int(w.left), Top: int(w.top), Right: int(w.right), Bottom: int(w.bottom)}, true
}

// fenced answers proposed moved the least distance that keeps it wholly inside work, its size kept.
func fenced(proposed rect, work placement.Rect) rect {
	size := placement.Size{Width: int(proposed.right - proposed.left), Height: int(proposed.bottom - proposed.top)}
	at := placement.Clamp(placement.Point{X: int(proposed.left), Y: int(proposed.top)}, size, work)
	return rect{left: int32(at.X), top: int32(at.Y), right: int32(at.X + size.Width), bottom: int32(at.Y + size.Height)}
}
