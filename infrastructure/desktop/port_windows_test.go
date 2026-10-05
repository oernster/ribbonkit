package desktop

import (
	"errors"
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// absentClass is a window class no window has, so finding it is refused.
const absentClass = "TimeRibbon port test, no such class"

// The port answers what the package's own functions answer, on a hidden window that is never shown.
// Opening the browser is asked with an address Windows has nothing to open with, so no browser opens.
func TestThePortIsThePackagesOwnOperations(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	port, ribbon := &Desktop{}, testRibbonWindow(t)
	at, size := placement.Point{X: 10, Y: 20}, placement.Size{Width: 120, Height: 80}
	if err := port.Place(ribbon, at, size); err != nil {
		t.Fatalf("placing through the port: %v", err)
	}
	got, err := port.Position(ribbon)
	if direct, _ := Position(ribbon); err != nil || got != direct || got != at {
		t.Errorf("the port reads %v (%v), the package %v; placed at %v", got, err, direct, at)
	}
	if err := port.Shape(ribbon, []placement.Rect{{Right: size.Width, Bottom: size.Height}}); err != nil {
		t.Errorf("shaping through the port: %v", err)
	}
	if err := port.SetTabFrame(ribbon, true); err != nil {
		t.Errorf("framing through the port: %v", err)
	}
	// Wails gives the ribbon WS_EX_APPWINDOW, which HideFromTaskbar relies on; the test window is given
	// it too, then must have it taken off for the tool window's style.
	_, _, _ = procSetWindowLongPtr.Call(uintptr(ribbon), uintptr(exStyleIndex), wsExAppWindow)
	if err := port.HideFromTaskbar(ribbon); err != nil {
		t.Errorf("hiding from the taskbar through the port: %v", err)
	}
	if style, _, _ := procGetWindowLongPtr.Call(uintptr(ribbon), uintptr(exStyleIndex)); style&wsExAppWindow != 0 || style&wsExToolWindow == 0 {
		t.Errorf("the ribbon's extended style is %#x after hiding it from the taskbar", style)
	}
	if err := port.KeepOnDisplays(ribbon, io.Discard); err != nil {
		t.Errorf("fencing through the port: %v", err)
	}
	if _, err := port.FindRibbon(absentClass, ""); !errors.Is(err, ErrRibbonNotFound) {
		t.Errorf("finding an absent class answered %v", err)
	}
	if err := port.OpenInBrowser(filepath.Join(t.TempDir(), "absent", "nothing-here.xyz")); err == nil {
		t.Error("an address with nothing to open it was reported opened")
	}
	readings := port.DragThreshold() == DragThreshold() && port.ToolkitScale() == ToolkitScale() &&
		port.PixelsPerDIP(1.5, ToolkitScale()) == PixelsPerDIP(1.5, ToolkitScale())
	if _, ok := port.Cursor(); !readings || !ok {
		t.Errorf("the port's readings differ from the package's (%v) or the cursor was unread", readings)
	}
}
