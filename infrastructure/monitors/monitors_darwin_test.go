package monitors

import (
	"os"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/cocoamain"
)

func TestMain(m *testing.M) { os.Exit(cocoamain.Serve(m.Run)) }

// A Cocoa rectangle is turned over against the menu-bar display: measured 2026-09-28 on a MacBook
// Air, a 1470x956 display whose visible frame is {0,75 1470x848}, below a 33 point menu bar and
// above a 75 point Dock.
func TestACocoaRectangleIsCountedFromTheTop(t *testing.T) {
	t.Parallel()
	got := turnedOver(cocoaRect{x: 0, y: 75, width: 1470, height: 848}, 956)
	want := placement.Rect{Left: 0, Top: 33, Right: 1470, Bottom: 881}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	above := turnedOver(cocoaRect{x: 200, y: 956, width: 1920, height: 1080}, 956)
	if above.Bottom != 0 || above.Top != -1080 {
		t.Errorf("a display above the menu-bar display was turned over to %+v", above)
	}
}

func TestADisplayIsNamedByPlaceAndName(t *testing.T) {
	t.Parallel()
	if got := deviceName(0, "Built-in Retina Display"); got != "display 1 Built-in Retina Display" {
		t.Errorf("got %q", got)
	}
	if got := deviceName(1, ""); got != "display 2" {
		t.Errorf("got %q", got)
	}
}

// The real displays: the first is the primary, every work area lies inside its display, every DPI
// is BaseDPI. The frames are logged for a person to check against System Settings.
func TestTheDisplaysAreReadInPoints(t *testing.T) {
	displays, err := readDisplays()
	if err != nil {
		t.Fatal(err)
	}
	for index, each := range displays {
		t.Logf("%s: work %+v frame %+v scale %v", each.monitor.Device, each.monitor.Work, each.geometry, each.scale)
		if each.monitor.Primary != (index == 0) {
			t.Errorf("%s: primary %v", each.monitor.Device, each.monitor.Primary)
		}
		if each.monitor.DPI != placement.BaseDPI {
			t.Errorf("%s: DPI %d", each.monitor.Device, each.monitor.DPI)
		}
		work, frame := each.monitor.Work, each.geometry
		if work.Left < frame.Left || work.Top < frame.Top || work.Right > frame.Right || work.Bottom > frame.Bottom {
			t.Errorf("%s: work area %+v outside its display %+v", each.monitor.Device, work, frame)
		}
	}
	if len(displays) > 0 && displays[0].geometry.Top != 0 {
		t.Errorf("the menu-bar display's top is %d, not 0", displays[0].geometry.Top)
	}
	public, err := Monitors{}.Monitors()
	if err != nil || len(public) != len(displays) {
		t.Errorf("Monitors answered %d displays (%v), read %d", len(public), err, len(displays))
	}
}
