package desktop

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
)

// wailsStyle is the style Wails gives its frameless window (measured 2026-09-28).
const wailsStyle = 0x04CA0000

// testRibbonWindow makes a hidden window styled as Wails styles the ribbon's, destroyed when the
// test ends. It is never shown. The caller locks its OS thread first, since resizing or restyling a
// window sends it messages its own thread must answer.
func testRibbonWindow(t *testing.T) Window {
	t.Helper()
	handle, _, err := procCreateWindowEx.Call(0, utf16Pointer("STATIC"), utf16Pointer("TimeRibbon test"), wailsStyle,
		0, 0, 200, 300, 0, 0, 0, 0)
	if handle == 0 {
		t.Fatalf("creating the test window: %v", err)
	}
	t.Cleanup(func() { _, _, _ = procDestroyWindow.Call(handle) })
	return Window(handle)
}

// FR-614: the tab's frame takes Wails' caption styles off and gives them back exactly. That the Wails
// window then takes 8 was measured on the probe (REQUIREMENTS section 2.3); a STATIC window stands in
// here and was measured to size otherwise (32 without them), so the width is left to a check by hand.
func TestTheTabFrameTakesTheCaptionStylesOffAndGivesThemBack(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window := testRibbonWindow(t)
	if err := SetTabFrame(window, true); err != nil {
		t.Fatal(err)
	}
	if style, _, _ := procGetWindowLongPtr.Call(uintptr(window), uintptr(styleIndex)); style != wailsStyle&^tabStyles|wsPopup {
		t.Errorf("as the tab, the style is %#x, want %#x", style, wailsStyle&^tabStyles|wsPopup)
	}
	// As a popup without those styles the window takes 8 either way, which a horizontal tab needs.
	for _, want := range []placement.Size{{Width: placement.TabThickness, Height: 300}, {Width: 300, Height: placement.TabThickness}} {
		if err := Place(window, placement.Point{}, want); err != nil {
			t.Fatal(err)
		}
		var bounds rect
		_, _, _ = procGetWindowRect.Call(uintptr(window), uintptr(unsafe.Pointer(&bounds)))
		if got := (placement.Size{Width: int(bounds.right - bounds.left), Height: int(bounds.bottom - bounds.top)}); got != want {
			t.Errorf("as the tab, asked %+v, took %+v", want, got)
		}
	}
	if err := SetTabFrame(window, true); err != nil {
		t.Errorf("taking the styles off twice: %v", err)
	}
	if err := SetTabFrame(window, false); err != nil {
		t.Fatal(err)
	}
	if style, _, _ := procGetWindowLongPtr.Call(uintptr(window), uintptr(styleIndex)); style != wailsStyle {
		t.Errorf("given back, the style is %#x, want %#x", style, wailsStyle)
	}
	if err := SetTabFrame(0, true); err == nil {
		t.Error("changing the frame of no window was not refused")
	}
}

// aroundPointer places window over the pointer where it stands now; else well away from it.
func aroundPointer(t *testing.T, window Window, over bool) {
	t.Helper()
	var cursor point
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	at := placement.Point{X: int(cursor.x) - 50, Y: int(cursor.y) - 50}
	if !over {
		at.X += 5000
	}
	if err := Place(window, at, placement.Size{Width: 100, Height: 100}); err != nil {
		t.Fatal(err)
	}
}

// FR-615, FR-616: the pointer is read against the window's rectangle; the pointer is never moved.
func TestThePointerIsReadAgainstTheWindow(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window := testRibbonWindow(t)
	aroundPointer(t, window, true)
	if inside, err := pointerInside(window); err != nil || !inside {
		t.Errorf("over the pointer: inside %v, %v", inside, err)
	}
	aroundPointer(t, window, false)
	if inside, err := pointerInside(window); err != nil || inside {
		t.Errorf("away from the pointer: inside %v, %v", inside, err)
	}
	if _, err := pointerInside(0); err == nil {
		t.Error("reading the pointer against no window was not refused")
	}
}

// FR-616, FR-913: the pointer is read against the window's shape, not its rectangle. A vertical
// ribbon's pull out is shorter than the ribbon, so the window is cut away above and below it;
// measured 2026-09-29, a pointer resting there counted as on the ribbon, which then never collapsed.
func TestThePointerIsReadAgainstTheWindowsShape(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window := testRibbonWindow(t)
	aroundPointer(t, window, true)
	// The pointer stands at the middle of the window, 50 in from its corner; this cut keeps a
	// corner square alone, well clear of it.
	corner := placement.Rect{Right: 20, Bottom: 20}
	if err := Shape(window, []placement.Rect{corner}); err != nil {
		t.Fatal(err)
	}
	if inside, err := pointerInside(window); err != nil || inside {
		t.Errorf("over the cut away part of the window: inside %v, %v", inside, err)
	}
	whole := placement.Rect{Right: 100, Bottom: 100}
	if err := Shape(window, []placement.Rect{whole}); err != nil {
		t.Fatal(err)
	}
	if inside, err := pointerInside(window); err != nil || !inside {
		t.Errorf("over the window's shape: inside %v, %v", inside, err)
	}
}

// Tracking reports where the pointer is as it starts, stops when asked and logs a window it cannot
// read once rather than at every reading.
func TestTrackingReportsThePointerAndStops(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window := testRibbonWindow(t)
	aroundPointer(t, window, true)
	log := &bytes.Buffer{}
	d := New(testApp, nil, log)
	d.TrackPointer(window, true)
	d.TrackPointer(window, true)
	select {
	case event := <-d.Events():
		if event.Kind != EventPointerArrived {
			t.Errorf("first report %d, want the pointer arrived", event.Kind)
		}
	case <-time.After(time.Second):
		t.Fatal("tracking reported nothing")
	}
	d.TrackPointer(window, false)
	d.TrackPointer(window, false)
	d.TrackPointer(0, true)
	if d.pointer.stop != nil {
		t.Error("tracking no window started a reading")
	}

	gone := &bytes.Buffer{}
	lost := New(testApp, nil, gone)
	lost.TrackPointer(Window(^uintptr(0)>>1), true)
	time.Sleep(3 * pointerEvery)
	lost.TrackPointer(0, false)
	if count := strings.Count(gone.String(), "reading the pointer"); count != 1 {
		t.Errorf("a window that cannot be read was logged %d times, want once: %q", count, gone.String())
	}
}
