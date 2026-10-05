package placement

import "testing"

// FR-623: while the grip is dragged the pull out keeps its size and its place along the ribbon from
// the ribbon's top-left corner, adjoining the same side of the ribbon as it grows; with the ribbon's
// corner still, the pull out's top-left corner never moves on the side the corner faces.
func TestThePullOutIsHeldWhileTheRibbonIsResized(t *testing.T) {
	t.Parallel()
	before := Rect{Left: 1000, Top: 100, Right: 1200, Bottom: 700}
	after := Rect{Left: 1000, Top: 100, Right: 1250, Bottom: 850}
	cases := []struct {
		side Edge
		held Rect
		want Rect
	}{
		{Left, Rect{Left: 600, Top: 300, Right: 1000, Bottom: 500}, Rect{Left: 600, Top: 300, Right: 1000, Bottom: 500}},
		{Right, Rect{Left: 1200, Top: 300, Right: 1600, Bottom: 500}, Rect{Left: 1250, Top: 300, Right: 1650, Bottom: 500}},
		{Top, Rect{Left: 900, Top: -100, Right: 1300, Bottom: 100}, Rect{Left: 900, Top: -100, Right: 1300, Bottom: 100}},
		{Bottom, Rect{Left: 900, Top: 700, Right: 1300, Bottom: 900}, Rect{Left: 900, Top: 850, Right: 1300, Bottom: 1050}},
	}
	for _, each := range cases {
		if got := PullOutHeld(after, each.side, each.held, before); got != each.want {
			t.Errorf("%s: %+v, want %+v", each.side, got, each.want)
		}
	}
}
