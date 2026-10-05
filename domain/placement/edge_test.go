package placement

import "testing"

var (
	tallRibbon = Size{Width: 176, Height: 400}
	wideRibbon = Size{Width: 600, Height: 92}
)

// FR-619, Oliver's report of 2026-09-29: a vertical ribbon counts as flush only against the left or
// right edge; one pushed against the top is flush against nothing. Horizontal likewise.
func TestOnlyAnEdgeAlongTheOrientationIsFlush(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		at       Point
		size     Size
		vertical bool
		want     Edge
		flush    bool
	}{
		{"vertical right", Point{X: 1920 - 176, Y: 300}, tallRibbon, true, Right, true},
		{"vertical left", Point{X: 0, Y: 300}, tallRibbon, true, Left, true},
		{"vertical at the top", Point{X: 800, Y: 0}, tallRibbon, true, "", false},
		{"vertical at the bottom", Point{X: 800, Y: 1032 - 400}, tallRibbon, true, "", false},
		{"vertical in the middle", Point{X: 800, Y: 300}, tallRibbon, true, "", false},
		{"horizontal top", Point{X: 700, Y: 0}, wideRibbon, false, Top, true},
		{"horizontal bottom", Point{X: 700, Y: 1032 - 92}, wideRibbon, false, Bottom, true},
		{"horizontal at the right", Point{X: 1920 - 600, Y: 400}, wideRibbon, false, "", false},
		{"horizontal at the left", Point{X: 0, Y: 400}, wideRibbon, false, "", false},
		{"hanging past the right", Point{X: 1920 - 170, Y: 300}, tallRibbon, true, "", false},
	}
	for _, each := range cases {
		edge, flush := FlushAgainst(each.at, each.size, primary.Work, each.vertical)
		if edge != each.want || flush != each.flush {
			t.Errorf("%s: got %q %v, want %q %v", each.name, edge, flush, each.want, each.flush)
		}
	}
}

// FR-619: every display's own work area counts, so the edge the primary shares with the display on
// its right is an edge of each.
func TestAnInnerEdgeCounts(t *testing.T) {
	t.Parallel()
	if edge, flush := FlushAgainst(Point{X: 1920 - 176, Y: 300}, tallRibbon, primary.Work, true); edge != Right || !flush {
		t.Errorf("against the primary's right: %q %v", edge, flush)
	}
	if edge, flush := FlushAgainst(Point{X: 1920, Y: 300}, tallRibbon, secondary.Work, true); edge != Left || !flush {
		t.Errorf("against the secondary's left: %q %v", edge, flush)
	}
}

// FR-410, its acceptance: 10 short of the right edge snaps flush, the top kept; 20 short stays.
// 10 past a shared edge also snaps back inside; an edge across the orientation never snaps.
func TestADropNearAnEdgeSnapsFlush(t *testing.T) {
	t.Parallel()
	reach := SnapReach
	cases := []struct {
		name     string
		at       Point
		size     Size
		vertical bool
		want     Point
	}{
		{"10 short of the right", Point{X: 1920 - 176 - 10, Y: 300}, tallRibbon, true, Point{X: 1920 - 176, Y: 300}},
		{"20 short of the right", Point{X: 1920 - 176 - 20, Y: 300}, tallRibbon, true, Point{X: 1920 - 176 - 20, Y: 300}},
		{"10 past the right", Point{X: 1920 - 176 + 10, Y: 300}, tallRibbon, true, Point{X: 1920 - 176, Y: 300}},
		{"10 from the left", Point{X: 10, Y: 300}, tallRibbon, true, Point{X: 0, Y: 300}},
		{"vertical near the top", Point{X: 800, Y: 5}, tallRibbon, true, Point{X: 800, Y: 5}},
		{"horizontal near the top", Point{X: 700, Y: 12}, wideRibbon, false, Point{X: 700, Y: 0}},
		{"horizontal near the right", Point{X: 1920 - 600 - 5, Y: 400}, wideRibbon, false, Point{X: 1920 - 600 - 5, Y: 400}},
		{"horizontal near the bottom", Point{X: 700, Y: 1032 - 92 - 16}, wideRibbon, false, Point{X: 700, Y: 1032 - 92}},
	}
	for _, each := range cases {
		if got := Snapped(each.at, each.size, primary.Work, each.vertical, reach); got != each.want {
			t.Errorf("%s: got %+v, want %+v", each.name, got, each.want)
		}
	}
}

// FR-410: where both edges are within reach, the nearer wins; at an equal distance, the near edge.
func TestTheNearerEdgeWinsWhenBothAreInReach(t *testing.T) {
	t.Parallel()
	narrow := Rect{Left: 0, Top: 0, Right: 200, Bottom: 1000}
	if got := Snapped(Point{X: 6, Y: 0}, tallRibbon, narrow, true, SnapReach); got.X != 0 {
		t.Errorf("6 from the left, 18 from the right: got %+v", got)
	}
	if got := Snapped(Point{X: 14, Y: 0}, tallRibbon, narrow, true, SnapReach); got.X != 200-176 {
		t.Errorf("14 from the left, 10 from the right: got %+v", got)
	}
}

// FR-410, FR-619: the edges that run along each orientation.
func TestTheEdgesAlongEachOrientation(t *testing.T) {
	t.Parallel()
	for edge, vertical := range map[Edge]bool{Left: true, Right: true, Top: false, Bottom: false} {
		if !Along(edge, vertical) || Along(edge, !vertical) {
			t.Errorf("%s runs along vertical %v only", edge, vertical)
		}
	}
}
