package arranger

import (
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// FR-902, FR-904: a horizontal ribbon at the top edge, its pull out on and open, has the pull out
// below it, 480 by 240 for its 336 long ribbon, centred on it; with none beside it, no side and no
// rectangle.
func TestAHorizontalRibbonsPullOutGoesBelowIt(t *testing.T) {
	t.Parallel()
	got, err := newRig(horizontal(), pulledOut(2)).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	depth := digitalCell.Height + 2*testPadding + testLane
	want := placement.Rect{Left: got.At.X + (336-480)/2, Top: depth, Right: got.At.X + (336-480)/2 + 480, Bottom: depth + 240}
	if got.PullOutSide != placement.Bottom || got.PullOut != want {
		t.Errorf("got side %s pull out %+v, want bottom %+v", got.PullOutSide, got.PullOut, want)
	}
	if off, _ := newRig(horizontal(), cells(2)).arranger.Launch(); off.PullOutSide != "" || off.PullOut != (placement.Rect{}) {
		t.Errorf("off: %+v", off)
	}
}

// FR-903, Amendment 22: a horizontal ribbon at the top edge has its handle below it; its pull out
// shows there only while it is open.
func TestAHorizontalRibbonsPullOutWaitsUntilOpen(t *testing.T) {
	t.Parallel()
	closed := pulledOut(2)
	closed.PullOut = false
	got, err := newRig(horizontal(), closed).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.PullOutSide != placement.Bottom || got.PullOut != (placement.Rect{}) {
		t.Errorf("closed: %+v", got)
	}
}

// FR-903, Amendment 23: the content's lane deepens the ribbon whichever way it runs, so a handle
// standing in it covers no cell; its length is unchanged.
func TestTheHandlesLaneDeepensTheRibbon(t *testing.T) {
	t.Parallel()
	for _, orientation := range []ribbon.Orientation{ribbon.Horizontal, ribbon.Vertical} {
		choices := ribbon.Defaults()
		choices.Orientation = orientation
		without, err := newRig(choices, cells(2)).arranger.Launch()
		if err != nil {
			t.Fatal(err)
		}
		with, err := newRig(choices, pulledOut(2)).arranger.Launch()
		if err != nil {
			t.Fatal(err)
		}
		wantDeeper := placement.Size{Width: without.Size.Width, Height: without.Size.Height + testLane}
		if orientation == ribbon.Vertical {
			wantDeeper = placement.Size{Width: without.Size.Width + testLane, Height: without.Size.Height}
		}
		if with.Size != wantDeeper {
			t.Errorf("%s: got %+v, want %+v", orientation, with.Size, wantDeeper)
		}
	}
}

// FR-902, Amendment 35, its acceptance: a vertical ribbon with its pull out on the left, dropped
// away from every edge where the right has more room, keeps it on the left, as it would carried onto
// another display; the side is kept for the next launch. Dropped where under 120 is left on its left,
// it goes right. Measured 2026-10-05: carried from the 100 percent display onto one at 250 percent,
// the clocks swapped sides.
func TestThePullOutKeepsItsSideAcrossADrop(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), pulledOut(2))
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	if side := r.host.current().PullOutSide; side != placement.Left {
		t.Fatalf("launched against the right edge the side kept is %q, want left", side)
	}
	got, err := r.arranger.Moved(placement.Point{X: 400, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.PullOutSide != placement.Left || got.PullOut.Right != got.At.X {
		t.Errorf("dropped with more room on the right the pull out went %q at %+v, want it kept on the left", got.PullOutSide, got.PullOut)
	}
	got, err = r.arranger.Moved(placement.Point{X: 60, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.PullOutSide != placement.Right || r.host.current().PullOutSide != placement.Right {
		t.Errorf("with 60 left on its left the pull out went %q, kept %q; want right", got.PullOutSide, r.host.current().PullOutSide)
	}
}

// FR-903: a vertical ribbon at the right edge has its handle on the left; its pull out shows there
// only while it is open.
func TestAVerticalRibbonsPullOutWaitsUntilOpen(t *testing.T) {
	t.Parallel()
	closed := pulledOut(2)
	closed.PullOut = false
	got, err := newRig(vertical(), closed).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.PullOutSide != placement.Left || got.PullOut != (placement.Rect{}) {
		t.Errorf("closed: %+v", got)
	}
	got, err = newRig(vertical(), pulledOut(2)).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.PullOutSide != placement.Left || got.PullOut.Right != got.At.X || got.PullOut.Width() != 480 {
		t.Errorf("open: %+v", got)
	}
}
