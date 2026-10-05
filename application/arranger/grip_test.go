package arranger

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// doubled is a scale that draws everything twice as large, so every length is exact.
const doubled = 2 * ribbon.WholeScale

// pulledOut answers n cells with a pull out beside them, open.
func pulledOut(n int) Content {
	content := cells(n)
	content.Lane, content.Beside, content.PullOut = testLane, true, true
	return content
}

// FR-623: at twice the scale a ribbon is twice as long and twice as thick, whichever way it runs;
// the scale is kept and drawn.
func TestAScaledRibbonGrowsInBothDirections(t *testing.T) {
	t.Parallel()
	for _, orientation := range []ribbon.Orientation{ribbon.Horizontal, ribbon.Vertical} {
		choices := ribbon.Defaults()
		choices.Orientation = orientation
		r := newRig(choices, cells(2))
		whole, err := r.arranger.Launch()
		if err != nil {
			t.Fatal(err)
		}
		if err := r.arranger.SetScale(doubled); err != nil {
			t.Fatal(err)
		}
		got, err := r.arranger.Launch()
		if err != nil {
			t.Fatal(err)
		}
		if want := (placement.Size{Width: 2 * whole.Size.Width, Height: 2 * whole.Size.Height}); got.Size != want {
			t.Errorf("%s: %+v, want %+v", orientation, got.Size, want)
		}
		kept := r.host.last(t).Scale
		if kept != doubled || r.arranger.DrawnScale(kept) != doubled {
			t.Errorf("%s: the scale was not kept and drawn", orientation)
		}
	}
}

// FR-623: while the grip is dragged the ribbon is drawn at the preview and nothing is saved; keeping
// a scale ends the preview; a scale outside the bounds is refused either way.
func TestAPreviewIsDrawnButNotKept(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	saves := r.host.saves()
	if err := r.arranger.PreviewScale(ribbon.MaxScale); err != nil {
		t.Fatal(err)
	}
	if got := r.arranger.DrawnScale(r.host.current().Scale); got != ribbon.MaxScale {
		t.Errorf("previewing draws %v", got)
	}
	if r.host.saves() != saves {
		t.Error("a preview was saved")
	}
	if err := r.arranger.SetScale(ribbon.MinScale); err != nil || r.arranger.DrawnScale(r.host.current().Scale) != ribbon.MinScale {
		t.Errorf("keeping a scale draws %v (%v)", r.arranger.DrawnScale(r.host.current().Scale), err)
	}
	for _, outside := range []int{ribbon.MinScale - 1, ribbon.MaxScale + 1} {
		if err := r.arranger.PreviewScale(float64(outside)); !errors.Is(err, ribbon.ErrUnknownChoice) {
			t.Errorf("previewing %d answered %v", outside, err)
		}
		if err := r.arranger.SetScale(outside); !errors.Is(err, ribbon.ErrUnknownChoice) || r.host.current().Scale != ribbon.MinScale {
			t.Errorf("keeping %d answered %v", outside, err)
		}
	}
}

// FR-623: a change of scale grows or shrinks the ribbon from its top-left corner, as a window being
// resized does, so the corner the grip is in follows the pointer; previewed or kept, it is never
// centred along its length again.
func TestAChangeOfScaleKeepsTheCorner(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1500, Y: 40}
	r := newRig(draggedTo(ribbon.Vertical, dragged), cells(2))
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	steps := []func() error{
		func() error { return r.arranger.PreviewScale(ribbon.WholeScale + ribbon.WholeScale/4) },
		func() error { return r.arranger.PreviewScale(ribbon.WholeScale + ribbon.WholeScale/2) },
		func() error { return r.arranger.SetScale(doubled) },
	}
	for index, step := range steps {
		if err := step(); err != nil {
			t.Fatal(err)
		}
		got, err := r.arranger.Rearrange(dragged)
		if err != nil {
			t.Fatal(err)
		}
		if got.At != dragged {
			t.Errorf("step %d: rearranged to %+v, want the corner kept at %+v", index, got.At, dragged)
		}
	}
}

// FR-623: while the grip is dragged the pull out keeps the size it had when the drag began, still
// adjoining the ribbon, so the window's corner holds still; once the scale is kept it is sized again.
func TestThePullOutIsHeldWhileTheGripIsDragged(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), pulledOut(6))
	before, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.arranger.PreviewScale(ribbon.WholeScale + ribbon.WholeScale/2); err != nil {
		t.Fatal(err)
	}
	during, err := r.arranger.Rearrange(before.At)
	if err != nil {
		t.Fatal(err)
	}
	if during.PullOut.Width() != before.PullOut.Width() || during.PullOut.Height() != before.PullOut.Height() || during.PullOut.Right != during.At.X {
		t.Errorf("during the drag the pull out is %+v beside a ribbon at %+v; want %+v's size adjoining it", during.PullOut, during.At, before.PullOut)
	}
	if err := r.arranger.SetScale(ribbon.WholeScale + ribbon.WholeScale/2); err != nil {
		t.Fatal(err)
	}
	kept, err := r.arranger.Rearrange(during.At)
	if err != nil {
		t.Fatal(err)
	}
	if kept.PullOut.Width() == before.PullOut.Width() {
		t.Errorf("once kept the pull out is still %d wide; want it sized for the new scale", kept.PullOut.Width())
	}
}

// FR-104: a change of content at the same scale still centres the ribbon along its new length.
func TestAChangeOfClocksAfterAScaleStillRecentres(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1500, Y: 40}
	r := newRig(draggedTo(ribbon.Vertical, dragged), cells(2))
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	if err := r.arranger.SetScale(doubled); err != nil {
		t.Fatal(err)
	}
	if _, err := r.arranger.Rearrange(dragged); err != nil {
		t.Fatal(err)
	}
	r.host.edit(func(c *Content) { c.Cells++ })
	got, err := r.arranger.Rearrange(dragged)
	if err != nil {
		t.Fatal(err)
	}
	if want := (primaryMonitor.Work.Height() - got.Size.Height) / 2; got.At.Y != want {
		t.Errorf("rearranged to %+v, want centred top to bottom at %d", got.At, want)
	}
}

// FR-623, FR-106: the scroll bar keeps the thickness the page measured whatever the scale, since it
// is the web engine's own; the cells beside it are scaled.
func TestTheScrollBarIsNotScaled(t *testing.T) {
	t.Parallel()
	const bar = 12
	crowded := horizontal()
	crowded.Scale = doubled
	r := newRig(crowded, cells(40))
	if err := r.arranger.SetScrollbar(bar); err != nil {
		t.Fatal(err)
	}
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := doubled*(digitalCell.Height+2*testPadding)/ribbon.WholeScale + bar
	if !got.Scrolls || got.Size.Height != want {
		t.Errorf("scrolls %v at %d thick, want %d", got.Scrolls, got.Size.Height, want)
	}
}
