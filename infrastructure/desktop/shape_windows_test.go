package desktop

import (
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// FR-913: the parts are joined into one region; no parts is no region, the whole window.
func TestThePartsMakeOneRegion(t *testing.T) {
	t.Parallel()
	if shape, err := regionOf(nil); shape != 0 || err != nil {
		t.Errorf("no parts: region %v (%v), want none", shape, err)
	}
	parts := []placement.Rect{{Left: 708, Right: 866, Bottom: 708}, {Top: 177, Right: 708, Bottom: 531}}
	shape, err := regionOf(parts)
	if shape == 0 || err != nil {
		t.Fatalf("two parts: region %v (%v)", shape, err)
	}
	procDeleteObject.Call(shape)
}

// FR-913: a window that is not there is refused with a reason, never ignored.
func TestShapingNoWindowIsRefused(t *testing.T) {
	t.Parallel()
	if err := Shape(0, []placement.Rect{{Right: 8, Bottom: 8}}); err == nil {
		t.Error("shaping no window answered no error")
	}
}
