package placement

import "testing"

// FR-614, its acceptance: a vertical ribbon 196 long flush against the right edge collapses to an
// 8 by 196 band flush against that edge; each edge takes the band on its own side.
func TestTheTabCoversTheFlushSide(t *testing.T) {
	t.Parallel()
	tall := Size{Width: 176, Height: 196}
	flush := Point{X: primary.Work.Right - tall.Width, Y: 400}
	if got, want := Tab(flush, tall, Right, TabThickness), (Rect{Left: 1912, Top: 400, Right: 1920, Bottom: 596}); got != want {
		t.Errorf("flush right: got %+v, want %+v", got, want)
	}
	if got, want := Tab(Point{X: 0, Y: 400}, tall, Left, TabThickness), (Rect{Left: 0, Top: 400, Right: 8, Bottom: 596}); got != want {
		t.Errorf("flush left: got %+v, want %+v", got, want)
	}
	wide := Size{Width: 600, Height: 92}
	if got, want := Tab(Point{X: 700, Y: 0}, wide, Top, TabThickness), (Rect{Left: 700, Top: 0, Right: 1300, Bottom: 8}); got != want {
		t.Errorf("flush top: got %+v, want %+v", got, want)
	}
	low := Point{X: 700, Y: primary.Work.Bottom - wide.Height}
	if got, want := Tab(low, wide, Bottom, TabThickness), (Rect{Left: 700, Top: 1024, Right: 1300, Bottom: 1032}); got != want {
		t.Errorf("flush bottom: got %+v, want %+v", got, want)
	}
}
