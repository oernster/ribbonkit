package desktop

import "testing"

// opaque and veiled are the opacities the veil asks the window manager for.
const (
	opaque = 1.0
	veiled = 0.0
)

// shownOpacityOf answers the opacity the window manager is asked to show ribbon at, skipping the
// test off X11, where the property the veil sets does not exist.
func shownOpacityOf(t *testing.T, ribbon Window) float64 {
	t.Helper()
	got, onX11, err := shownOpacity(ribbon)
	if err != nil {
		t.Fatal(err)
	}
	if !onX11 {
		t.Skip("the window manager's opacity is an X11 property; run with GDK_BACKEND=x11")
	}
	return got
}

// A window mapped before GTK's main loop runs is the one Wails shows too early: the window manager
// is asked to show it transparent, then opaque when it is next mapped, which is the ribbon's
// showing. It is the property the window manager reads that is checked, not GTK's own opacity,
// which a translucent window ignores until it paints.
func TestAWindowMappedBeforeTheLoopIsVeiledUntilItIsShown(t *testing.T) {
	veilEarlyMaps()
	ribbon := newTestWindow(testApp.Name)
	defer closeTestWindow(ribbon)
	if err := mapAtLevel(ribbon, 0); err != nil {
		t.Fatal(err)
	}
	if got := shownOpacityOf(t, ribbon); got != veiled {
		t.Errorf("mapped before the loop: opacity %v, want %v", got, veiled)
	}
	if err := mapAtLevel(ribbon, 1); err != nil {
		t.Fatal(err)
	}
	if got := shownOpacityOf(t, ribbon); got != opaque {
		t.Errorf("shown in the loop: opacity %v, want %v", got, opaque)
	}
}

// A window first mapped with the loop running was never veiled, so it is never made transparent.
func TestTheVeilLeavesAWindowItNeverVeiled(t *testing.T) {
	ribbon := newTestWindow(testApp.Name)
	defer closeTestWindow(ribbon)
	if err := mapAtLevel(ribbon, 1); err != nil {
		t.Fatal(err)
	}
	if got := shownOpacityOf(t, ribbon); got != opaque {
		t.Errorf("opacity %v, want %v", got, opaque)
	}
}
