package monitors

import (
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// Reads the real displays. What it can hold on any machine: at least one display, exactly one
// primary, a device name, a work area with area and a DPI no lower than 100 percent.
func TestTheDisplaysAreReadWithTheirIdentityAndWorkArea(t *testing.T) {
	found, err := Monitors{}.Monitors()
	if err != nil {
		t.Fatal(err)
	}
	if len(found) == 0 {
		t.Fatal("no displays")
	}
	primaries := 0
	for _, monitor := range found {
		t.Logf("%+v", monitor)
		if monitor.Primary {
			primaries++
		}
		if !strings.HasPrefix(monitor.Device, `\\.\`) {
			t.Errorf("device name %q", monitor.Device)
		}
		if monitor.Work.Width() <= 0 || monitor.Work.Height() <= 0 {
			t.Errorf("work area %+v", monitor.Work)
		}
		if monitor.DPI < placement.BaseDPI {
			t.Errorf("DPI %d", monitor.DPI)
		}
	}
	if primaries != 1 {
		t.Errorf("%d primaries", primaries)
	}
}

// A second enumeration answers the same displays, so the shared callback is reset between them.
func TestEnumeratingTwiceAnswersTheSameDisplays(t *testing.T) {
	first, err := Monitors{}.Monitors()
	if err != nil {
		t.Fatal(err)
	}
	second, err := Monitors{}.Monitors()
	if err != nil || len(second) != len(first) {
		t.Errorf("first %d, second %d (%v)", len(first), len(second), err)
	}
}
