package arranger

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// FR-614: the tab is the band on the side flush against the ribbon's edge, in that display's pixels.
// Collapsing saves nothing, so the stored place stays the full ribbon's.
func TestCollapsingKeepsThePlacement(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(1))
	full, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	saves := r.host.saves()
	tab, err := r.arranger.Collapsed(full)
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{
		At:   placement.Point{X: primaryMonitor.Work.Right - placement.TabThickness, Y: full.At.Y},
		Size: placement.Size{Width: placement.TabThickness, Height: full.Size.Height},
		DPI:  placement.BaseDPI,
	}
	if tab != want {
		t.Errorf("got %+v, want %+v", tab, want)
	}
	if r.host.saves() != saves {
		t.Error("collapsing saved the settings")
	}
	// At 150 percent the band is 12 pixels deep, on the side against that display's right edge.
	wide := Arrangement{At: placement.Point{X: 4000, Y: 100}, Size: placement.Size{Width: 264, Height: 300}, Edge: placement.Right}
	scaled, err := r.arranger.Collapsed(wide)
	if err != nil {
		t.Fatal(err)
	}
	if scaled.At.X != 4264-12 || scaled.Size.Width != 12 || scaled.DPI != secondaryMonitor.DPI {
		t.Errorf("at 150 percent: got %+v", scaled)
	}
}

func TestCollapsingWithoutDisplaysIsRefused(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(1))
	flush := Arrangement{Edge: placement.Right}
	r.monitors.monitors = nil
	if _, err := r.arranger.Collapsed(flush); !errors.Is(err, ErrNoMonitors) {
		t.Errorf("no monitors: got %v", err)
	}
	r.monitors.err = errPlanted
	if _, err := r.arranger.Collapsed(flush); !errors.Is(err, errPlanted) {
		t.Errorf("a fault: got %v", err)
	}
}

// FR-619: a ribbon flush against no edge never collapses, so its tab is refused.
func TestARibbonAgainstNoEdgeHasNoTab(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(1))
	if _, err := r.arranger.Collapsed(Arrangement{At: placement.Point{X: 800, Y: 300}}); !errors.Is(err, ErrNotAgainstAnEdge) {
		t.Errorf("got %v", err)
	}
}
