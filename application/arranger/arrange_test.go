package arranger

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// unchanged is an edit that changes nothing, which still saves.
func unchanged(current ribbon.Choices) ribbon.Choices { return current }

// FR-403, FR-409: two digital cells at 100 percent, 2 x 160 + 2 x 8 = 336 along and
// 90 + 2 x 8 = 106 across, go to their orientation's home edge: flush against the top, centred left
// to right, when horizontal; flush against the right, centred top to bottom, when vertical.
func TestLaunchWithNothingStoredGoesToTheDefaultPlace(t *testing.T) {
	t.Parallel()
	got, err := newRig(horizontal(), cells(2)).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{
		At:   placement.Point{X: (1920 - 336) / 2, Y: 0},
		Size: placement.Size{Width: 336, Height: 106}, DPI: placement.BaseDPI, Edge: placement.Top,
	}
	if got != want {
		t.Errorf("horizontal: got %+v, want %+v", got, want)
	}
	// Vertical, the same two cells are 160 + 2 x 8 = 176 across and 2 x 90 + 2 x 8 = 196 along.
	if got, _ := newRig(vertical(), cells(2)).arranger.Launch(); got.At != (placement.Point{X: 1920 - 176, Y: (1032 - 196) / 2}) {
		t.Errorf("vertical: got %+v", got)
	}
}

// FR-405, FR-407: the stored monitor at 150 percent sizes the ribbon in its pixels.
func TestLaunchRestoresTheStoredMonitorAtItsScaling(t *testing.T) {
	t.Parallel()
	initial := horizontal()
	initial.Placement = &placement.Stored{Device: secondaryMonitor.Device, DPI: 144, Offset: placement.Point{X: 180, Y: 300}}
	got, err := newRig(initial, cells(2)).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{At: placement.Point{X: 2100, Y: 300}, Size: placement.Size{Width: 504, Height: 159}, DPI: 144}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// FR-103: three analogue cells stacked: 160 + 16 across, 3 x 150 + 16 along.
func TestVerticalRibbonsStackTheirCells(t *testing.T) {
	t.Parallel()
	content := cells(3)
	content.Cell = analogueCell
	got, _ := newRig(vertical(), content).arranger.Launch()
	if got.Size != (placement.Size{Width: 176, Height: 466}) || got.Scrolls {
		t.Errorf("got %+v", got)
	}
}

// FR-104: a vertical ribbon dragged near the top gains a cell. Its length changes from
// 2 x 90 + 8 = 188 to 3 x 90 + 2 x 8 = 286, so it is centred top to bottom, (1032 - 286) / 2,
// its left edge kept; the place is saved, so the next launch finds it there.
func TestARibbonWhoseLengthChangesIsRecentredAndKept(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1700, Y: 40}
	r := newRig(draggedTo(ribbon.Vertical, dragged), cells(2))
	if got, _ := r.arranger.Launch(); got.At != dragged {
		t.Fatalf("launched at %+v, want where the drag left it", got.At)
	}
	r.host.edit(func(c *Content) { c.Cells++ })
	centred := placement.Point{X: 1700, Y: (1032 - 286) / 2}
	got, err := r.arranger.Rearrange(dragged)
	if err != nil || got.At != centred {
		t.Errorf("rearranged to %+v (%v), want %+v", got.At, err, centred)
	}
	if stored := r.host.last(t).Placement; stored == nil || stored.Offset != centred {
		t.Errorf("stored %+v, want the centred place", stored)
	}
	if got, _ := r.arranger.Launch(); got.At != centred {
		t.Errorf("relaunched at %+v, want the centred place", got.At)
	}
}

// FR-104: a horizontal ribbon is centred left to right, its top kept.
func TestAHorizontalRibbonIsRecentredLeftToRight(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 30, Y: 800}
	r := newRig(draggedTo(ribbon.Horizontal, dragged), cells(2))
	_, _ = r.arranger.Launch()
	r.host.edit(func(c *Content) { c.Cells++ })
	got, _ := r.arranger.Rearrange(dragged)
	if want := (placement.Point{X: (1920 - (3*160 + 2*8)) / 2, Y: 800}); got.At != want {
		t.Errorf("got %+v, want %+v", got.At, want)
	}
}

// FR-104, FR-404: only a change of length re-centres the ribbon; a rearrange or a move with the
// same cells leaves it where it was put. Nothing is saved but the move.
func TestNothingButAChangeOfLengthRecentresTheRibbon(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1700, Y: 40}
	r := newRig(draggedTo(ribbon.Vertical, dragged), cells(2))
	_, _ = r.arranger.Launch()
	if got, _ := r.arranger.Rearrange(dragged); got.At != dragged || r.host.saves() != 0 {
		t.Errorf("rearranged to %+v with %d saves, want it left alone", got.At, r.host.saves())
	}
	moved := placement.Point{X: 1700, Y: 500}
	if got, _ := r.arranger.Moved(moved); got.At != moved || r.host.last(t).Placement.Offset != moved {
		t.Errorf("a drag was moved to %+v, want it kept where it was let go", got.At)
	}
}

// FR-106.
func TestARibbonThatWillNotFitScrollsAtTheWidthOfTheWorkArea(t *testing.T) {
	t.Parallel()
	got, _ := newRig(horizontal(), cells(12)).arranger.Launch()
	if got.Size.Width != 1920 || !got.Scrolls || got.At.X != 0 {
		t.Errorf("got %+v", got)
	}
}

// FR-104, FR-707: a ribbon re-centred where its place cannot be saved has the host raise a notice,
// one more cell, so it is arranged once more with room for that cell rather than left too short
// for it: 3 x 160 + 2 x 8 = 496 along.
func TestARecentringThatCannotBeSavedMakesRoomForItsNotice(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	r.host.failSaves(errPlanted)
	_ = r.host.ChangeRibbon(unchanged)
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	r.host.dismiss()
	got, err := r.arranger.Launch()
	if err != nil || got.Size.Width != 496 {
		t.Errorf("got %+v (%v), want room for the notice the failed save raised", got, err)
	}
}

// FR-106: a ribbon that scrolls is made thicker by the scroll bar the page reports, so the bar never
// covers the cells; one that fits is not. 12 cells overflow the primary: 90 + 16 + 15 = 121 across.
func TestAScrollingRibbonMakesRoomForItsScrollBar(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(12))
	const bar = 15
	if err := r.arranger.SetScrollbar(bar); err != nil {
		t.Fatal(err)
	}
	got, _ := r.arranger.Launch()
	if !got.Scrolls || got.Size.Height != 90+2*testPadding+bar {
		t.Errorf("scrolling: got %+v", got)
	}
	fits := newRig(horizontal(), cells(2))
	if err := fits.arranger.SetScrollbar(bar); err != nil {
		t.Fatal(err)
	}
	if got, _ := fits.arranger.Launch(); got.Size.Height != 90+2*testPadding {
		t.Errorf("fitting: got %+v", got)
	}
	if err := r.arranger.SetScrollbar(-1); !errors.Is(err, placement.ErrNegativeLength) {
		t.Errorf("a negative bar: got %v", err)
	}
}

// FR-404.
func TestPlacementIsStoredRelativeToItsMonitor(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	got, err := r.arranger.Moved(placement.Point{X: 2100, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	stored := r.host.last(t).Placement
	want := placement.Stored{Device: secondaryMonitor.Device, Work: secondaryMonitor.Work, DPI: 144, Offset: placement.Point{X: 180, Y: 300}}
	if stored == nil || *stored != want || got.At != (placement.Point{X: 2100, Y: 300}) {
		t.Errorf("stored %+v at %+v", stored, got.At)
	}
}

// FR-406: a horizontal ribbon dragged off every display comes back to the default place, flush
// against the primary's top (FR-403, FR-409), 336 x 106; the place it comes back to is stored.
func TestADragOffEveryDisplayIsBroughtBack(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	got, err := r.arranger.Moved(placement.Point{X: 9000, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: (1920 - 336) / 2, Y: 0}) || r.host.last(t).Placement.Device != primaryMonitor.Device {
		t.Errorf("got %+v, stored %+v", got, r.host.last(t).Placement)
	}
}

// FR-104, FR-406: rearranging keeps the corner, clamps it and saves no placement. Clamped onto the
// bottom edge, the horizontal ribbon now stands flush against it, which is remembered (FR-411).
func TestRearrangingClampsAndSavesNothing(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	got, err := r.arranger.Rearrange(placement.Point{X: 1500, Y: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1500, Y: 1032 - 106}) || got.Edge != placement.Bottom {
		t.Errorf("got %+v", got)
	}
	if last := r.host.last(t); last.Placement != nil || last.LastEdge == nil || *last.LastEdge != (placement.Against{Device: primaryMonitor.Device, Edge: placement.Bottom}) {
		t.Errorf("rearranging saved placement %+v, edge %+v", last.Placement, last.LastEdge)
	}
}

// FR-407: at (1800, 1000) the ribbon overlaps the secondary most, so it is sized in the secondary's
// pixels (504 by 159 at 150 percent) and clamped onto it.
func TestARibbonLandingOnAnotherDisplayIsSizedForIt(t *testing.T) {
	t.Parallel()
	got, err := newRig(horizontal(), cells(2)).arranger.Rearrange(placement.Point{X: 1800, Y: 1000})
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{At: placement.Point{X: 1920, Y: 1000}, Size: placement.Size{Width: 504, Height: 159}, DPI: 144}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// CON-6: Settings opens centred on the ribbon's display, sized in its pixels.
func TestSettingsOpenCentredOnTheRibbonsDisplay(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(1))
	got, err := r.arranger.Centred(placement.Point{X: 2100, Y: 300}, placement.Size{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{At: placement.Point{X: 1920 + (2560-600)/2, Y: (1392 - 450) / 2}, Size: placement.Size{Width: 600, Height: 450}, DPI: 144}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if r.host.saves() != 0 {
		t.Error("centring saved something")
	}
	tall, _ := r.arranger.Centred(placement.Point{X: 10, Y: 10}, placement.Size{Width: 400, Height: 2000})
	if tall.Size.Height != 1032 || tall.At.Y != 0 {
		t.Errorf("a surface taller than the work area is not capped to it: %+v", tall)
	}
}

func TestNoDisplaysOrAFaultReadingThemIsAnswered(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(1))
	r.monitors.monitors = nil
	if _, err := r.arranger.Launch(); !errors.Is(err, ErrNoMonitors) {
		t.Errorf("no monitors: got %v", err)
	}
	r.monitors.err = errPlanted
	if _, err := r.arranger.Rearrange(placement.Point{}); !errors.Is(err, errPlanted) {
		t.Errorf("a fault: got %v", err)
	}
	if _, err := r.arranger.Moved(placement.Point{}); !errors.Is(err, errPlanted) {
		t.Errorf("a fault while moving: got %v", err)
	}
	if _, err := r.arranger.Centred(placement.Point{}, placement.Size{}); !errors.Is(err, errPlanted) {
		t.Errorf("a fault while centring: got %v", err)
	}
	r.monitors.err = nil
	if _, err := r.arranger.Centred(placement.Point{}, placement.Size{}); !errors.Is(err, ErrNoMonitors) {
		t.Errorf("centring with no monitors: got %v", err)
	}
}
