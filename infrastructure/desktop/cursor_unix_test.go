//go:build linux || darwin

package desktop

import (
	"testing"
	"time"

	"github.com/oernster/ribbonkit/domain/placement"
)

// FR-623: the pointer is read in the units the ribbon is placed in, counted from the same corner, so
// the corner grip's drag can follow the desktop's pointer rather than the page's. The pointer is moved
// onto a window of the process's own (an X11 window under XWayland), read back, then put back where
// it was; a person at the desktop sees it jump.
func TestThePointerIsReadWhereTheRibbonIsPlaced(t *testing.T) {
	from, ok := Cursor()
	if !ok {
		t.Fatal("the pointer could not be read")
	}
	defer warpPointer(from)
	ribbon := newTestWindow(testApp.Name)
	defer closeTestWindow(ribbon)
	work := primaryWorkArea(t)
	corner := placement.Point{X: work.Left + work.Width()/4, Y: work.Top + work.Height()/4}
	if err := Place(ribbon, corner, testSize); err != nil {
		t.Fatal(err)
	}
	target := placement.Point{X: corner.X + testSize.Width/2, Y: corner.Y + testSize.Height/2}
	var at placement.Point
	for deadline := time.Now().Add(settleLimit); time.Now().Before(deadline); time.Sleep(settlePause) {
		warpPointer(target)
		if at, ok = Cursor(); ok && at == target {
			break
		}
	}
	t.Logf("moved the pointer to %+v; read %+v", target, at)
	if !ok || at != target {
		t.Errorf("the pointer reads %+v, moved to %+v", at, target)
	}
}
