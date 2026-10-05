package monitors

import (
	"os"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/gtkmain"
)

func TestMain(m *testing.M) { os.Exit(gtkmain.ServeTests(m.Run)) }

// Reads the real displays. What it can hold on any machine: at least one display, exactly one
// primary, a device name, a work area with area inside the display, every display at BaseDPI. The
// geometry and scale factor are logged, since they are what settles the units.
func TestTheDisplaysAreReadWithTheirIdentityAndWorkArea(t *testing.T) {
	found, err := readDisplays()
	if err != nil {
		t.Fatal(err)
	}
	if len(found) == 0 {
		t.Fatal("no displays")
	}
	primaries := 0
	for _, each := range found {
		monitor := each.monitor
		t.Logf("%+v geometry %+v scale %d", monitor, each.geometry, each.scale)
		if monitor.Primary {
			primaries++
		}
		if !strings.HasPrefix(monitor.Device, "display ") {
			t.Errorf("device name %q", monitor.Device)
		}
		if monitor.Work.Width() <= 0 || monitor.Work.Height() <= 0 {
			t.Errorf("work area %+v", monitor.Work)
		}
		if monitor.Work.Left < each.geometry.Left || monitor.Work.Right > each.geometry.Right ||
			monitor.Work.Top < each.geometry.Top || monitor.Work.Bottom > each.geometry.Bottom {
			t.Errorf("work area %+v outside the display %+v", monitor.Work, each.geometry)
		}
		if monitor.DPI != placement.BaseDPI || each.scale < 1 {
			t.Errorf("DPI %d scale %d", monitor.DPI, each.scale)
		}
	}
	if primaries != 1 {
		t.Errorf("%d primaries", primaries)
	}
}

func TestADisplayIsNamedByItsPlaceAndModel(t *testing.T) {
	if got := deviceName(0, ""); got != "display 1" {
		t.Errorf("got %q", got)
	}
	if got := deviceName(1, "BOE 0x0BCA"); got != "display 2 BOE 0x0BCA" {
		t.Errorf("got %q", got)
	}
}
