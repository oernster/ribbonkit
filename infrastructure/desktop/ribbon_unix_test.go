//go:build linux || darwin

package desktop

import (
	"errors"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/monitors"
)

// The window manager acts on a placement when it gets to it, so the result is polled for.
const (
	settleLimit = 2 * time.Second
	settlePause = 20 * time.Millisecond
)

// The test ribbon's size, in DIP (GTK's units, AppKit's points): long and short, as a horizontal
// ribbon is.
var testSize = placement.Size{Width: 300, Height: 80}

func TestTheRibbonIsFoundByItsTitle(t *testing.T) {
	made := newTestWindow(testApp.Name)
	defer closeTestWindow(made)
	found, err := FindRibbon("", testApp.Name)
	if err != nil || found != made {
		t.Errorf("found %d (%v), made %d", found, err, made)
	}
}

// FR-405: the ribbon stands where it is placed, at the size it is given. This is the measurement
// the Linux design (on X11) and the macOS design both rest on.
func TestTheRibbonGoesWhereItIsPlaced(t *testing.T) {
	ribbon := newTestWindow(testApp.Name)
	defer closeTestWindow(ribbon)
	work := primaryWorkArea(t)
	target := placement.Point{X: work.Left + work.Width()/4, Y: work.Top + work.Height()/4}
	if err := Place(ribbon, target, testSize); err != nil {
		t.Fatal(err)
	}
	var at placement.Point
	var got placement.Size
	for deadline := time.Now().Add(settleLimit); time.Now().Before(deadline); time.Sleep(settlePause) {
		at, _ = Position(ribbon)
		got, _ = size(ribbon)
		if at == target && got == testSize {
			break
		}
	}
	t.Logf("placed at %+v size %+v; stands at %+v size %+v", target, testSize, at, got)
	if at != target || got != testSize {
		t.Errorf("the ribbon stands at %+v size %+v, not where it was placed", at, got)
	}
}

// CON-6, FR-405: closing a panel shrinks the window back to the ribbon where it was left. Measured
// 2026-09-28 in the running app: after a 560x760 panel, a ribbon placed at (1252,358) stood at
// (880,152) instead.
func TestTheRibbonReturnsFromAPanelToWhereItIsPlaced(t *testing.T) {
	ribbon := newTestWindow(testApp.Name)
	defer closeTestWindow(ribbon)
	work := primaryWorkArea(t)
	ribbonAt := placement.Point{X: work.Right - testSize.Width, Y: work.Top + work.Height()/3}
	panelSize := placement.Size{Width: 560, Height: 760}
	panelAt := placement.Point{X: work.Left + (work.Width()-panelSize.Width)/2, Y: work.Top + (work.Height()-panelSize.Height)/2}
	for _, step := range []struct {
		at   placement.Point
		size placement.Size
	}{{ribbonAt, testSize}, {panelAt, panelSize}, {ribbonAt, testSize}} {
		if err := Place(ribbon, step.at, step.size); err != nil {
			t.Fatal(err)
		}
		at, got := settled(ribbon, step.at, step.size)
		t.Logf("placed at %+v size %+v; stands at %+v size %+v", step.at, step.size, at, got)
		if at != step.at || got != step.size {
			t.Errorf("stands at %+v size %+v, not where it was placed", at, got)
		}
	}
}

// settled answers where ribbon stands once it has reached at and size; failing that, once
// settleLimit passes.
func settled(ribbon Window, at placement.Point, want placement.Size) (placement.Point, placement.Size) {
	var stands placement.Point
	var got placement.Size
	for deadline := time.Now().Add(settleLimit); time.Now().Before(deadline); time.Sleep(settlePause) {
		stands, _ = Position(ribbon)
		got, _ = size(ribbon)
		if stands == at && got == want {
			break
		}
	}
	return stands, got
}

// FR-401: the threshold is the desktop's own, which is never zero.
func TestTheDragThresholdIsTheDesktopsOwn(t *testing.T) {
	got := DragThreshold()
	if got.Width <= 0 || got.Width != got.Height {
		t.Errorf("got %+v", got)
	}
	t.Logf("drag threshold %+v DIP", got)
}

// A Window this package never handed out is refused rather than acted on.
func TestAnUnknownWindowIsRefused(t *testing.T) {
	if err := Place(Window(^uintptr(0)), placement.Point{}, testSize); !errors.Is(err, errUnknownWindow) {
		t.Errorf("got %v", err)
	}
}

// primaryWorkArea answers the primary display's work area.
func primaryWorkArea(t *testing.T) placement.Rect {
	t.Helper()
	found, err := monitors.Monitors{}.Monitors()
	if err != nil {
		t.Fatal(err)
	}
	for _, monitor := range found {
		if monitor.Primary {
			return monitor.Work
		}
	}
	t.Fatal("no primary display")
	return placement.Rect{}
}
