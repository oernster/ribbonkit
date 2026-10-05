package window

import (
	"math"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// The page's background reaches the window, so a window catching up with a new size shows it rather
// than white (measured 2026-09-29); opaque while the window is wholly opaque.
func TestThePagesBackgroundReachesTheWindow(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.choices.Opacity = ribbon.MaxOpacity
	if err := app.SetBackground(7, 36, math.MaxUint8); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(seen.backgrounds, [][4]uint8{{7, 36, math.MaxUint8, math.MaxUint8}}) {
		t.Errorf("painted %v", seen.backgrounds)
	}
}

// FR-622: below full opacity the window's own paint is clear, so the desktop shows through the page;
// a change of opacity paints again at once in the colour last reported, before any report nothing.
func TestTheWindowIsPaintedClearBelowFullOpacity(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.choices.Opacity = ribbon.MaxOpacity
	if err := app.SetOpacity(ribbon.MinOpacity); err != nil || len(seen.backgrounds) != 0 {
		t.Fatalf("before the page reported a colour: %v, painted %v", err, seen.backgrounds)
	}
	if err := app.SetBackground(7, 36, 9); err != nil {
		t.Fatal(err)
	}
	if err := app.SetOpacity(ribbon.MaxOpacity); err != nil {
		t.Fatal(err)
	}
	want := [][4]uint8{{7, 36, 9, 0}, {7, 36, 9, math.MaxUint8}}
	if !slices.Equal(seen.backgrounds, want) || !slices.Contains(service.calls, "SetOpacity") {
		t.Errorf("painted %v, want %v", seen.backgrounds, want)
	}
}

// The page is foreign input: a channel outside a byte is refused, never wrapped into another colour.
func TestABackgroundThatIsNotAColourIsRefused(t *testing.T) {
	t.Parallel()
	app, _, seen, _ := newTestApp(t)
	for _, channel := range []int{-1, math.MaxUint8 + 1} {
		if err := app.SetBackground(0, channel, 0); err == nil {
			t.Errorf("channel %d was taken", channel)
		}
	}
	if len(seen.backgrounds) != 0 {
		t.Errorf("painted %v", seen.backgrounds)
	}
}
