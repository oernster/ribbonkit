package desktop

import "testing"

// The window takes the page's ratio over GTK's own window scale: KDE's fractional scale arrives as
// font DPI alone; GTK's whole scale is already in its units; a scale below one counts as none.
func TestTheWindowTakesThePagesRatioOverGTKsScale(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		name    string
		ratio   float64
		toolkit int
		want    float64
	}{
		{"unscaled", 1, 1, 1},
		{"fractional scale as font DPI", 1.5, 1, 1.5},
		{"whole scale GTK applies", 2, 2, 1},
		{"whole scale with font DPI on top", 3, 2, 1.5},
		{"no scale reported", 1.5, 0, 1.5},
	} {
		if got := PixelsPerDIP(each.ratio, each.toolkit); got != each.want {
			t.Errorf("%s: ratio %v over GTK scale %d gave %v, want %v", each.name, each.ratio, each.toolkit, got, each.want)
		}
	}
}

// GDK is asked on its own loop and answers a usable whole scale.
func TestGTKsWindowScaleIsRead(t *testing.T) {
	if got := ToolkitScale(); got < wholeScale {
		t.Errorf("GTK's window scale read as %d", got)
	}
}
