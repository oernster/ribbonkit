package placement

import "testing"

// timeRibbon is FR-412's worked example: a vertical ribbon 176 by 196 against the right edge of work,
// centred, from (1744, 418) to (1920, 614).
var timeRibbon = Rect{Left: 1744, Top: 418, Right: 1920, Bottom: 614}

// alone answers a footprint of the ribbon itself, size across, with no pull out.
func alone(size Size) func(Point) []Rect {
	return func(at Point) []Rect { return []Rect{rectOf(at, size)} }
}

// FR-412, its acceptance: a ribbon 300 tall wanting the centre of the right edge, from 366 to 666,
// ties at 248 between above and below the ribbon there and stands above it; dropped with its top at
// 500 it stands below, 114 away rather than 382.
func TestARibbonStandsAtTheNearestClearPlaceAlongItsEdge(t *testing.T) {
	t.Parallel()
	size := Size{Width: 176, Height: 300}
	cases := []struct {
		name   string
		wanted Point
		want   Point
	}{
		{"centred, a tie", Point{X: 1744, Y: 366}, Point{X: 1744, Y: 118}},
		{"dropped lower", Point{X: 1744, Y: 500}, Point{X: 1744, Y: 614}},
		{"already clear", Point{X: 1744, Y: 0}, Point{X: 1744, Y: 0}},
	}
	for _, each := range cases {
		got, ok := Clear(each.wanted, size, work, true, []Rect{timeRibbon}, alone(size))
		if !ok || got != each.want {
			t.Errorf("%s: got %+v %v, want %+v", each.name, got, ok, each.want)
		}
	}
}

// FR-412: touching is not overlapping, so a ribbon ending where another begins stays where it is.
func TestTouchingIsNotOverlapping(t *testing.T) {
	t.Parallel()
	size := Size{Width: 176, Height: 300}
	if got, ok := Clear(Point{X: 1744, Y: 118}, size, work, true, []Rect{timeRibbon}, alone(size)); !ok || got.Y != 118 {
		t.Errorf("got %+v %v, want it kept at 118", got, ok)
	}
}

// FR-412, its acceptance: a ribbon 900 tall finds no clear 900 on the right edge, 418 above and 418
// below; Clear says so rather than overlapping.
func TestNoClearPlaceIsSaid(t *testing.T) {
	t.Parallel()
	size := Size{Width: 176, Height: 900}
	if got, ok := Clear(Point{X: 1744, Y: 66}, size, work, true, []Rect{timeRibbon}, alone(size)); ok {
		t.Errorf("got %+v, want no clear place", got)
	}
}

// FR-412, its acceptance: a ribbon against no edge keeps off another's shown pull out, sliding along
// its own length: dropped at (1400, 300) beside a sun map from (1264, 396) to (1744, 636) it stands
// at (1400, 96), 204 up rather than 336 down.
func TestARibbonKeepsOffAnotherPullOut(t *testing.T) {
	t.Parallel()
	size := Size{Width: 176, Height: 300}
	sunMap := Rect{Left: 1264, Top: 396, Right: 1744, Bottom: 636}
	got, ok := Clear(Point{X: 1400, Y: 300}, size, work, true, []Rect{timeRibbon, sunMap}, alone(size))
	if !ok || got != (Point{X: 1400, Y: 96}) {
		t.Errorf("got %+v %v, want (1400, 96)", got, ok)
	}
}

// FR-412, OQ-38: the ribbon's own pull out counts. A horizontal ribbon at the top whose 480 by 240 pull
// out hangs below it, centred, would clear another ribbon with its own body yet cover it with its
// pull out, so it moves along until the pull out clears too.
func TestTheRibbonsOwnPullOutKeepsOffTheOther(t *testing.T) {
	t.Parallel()
	size := Size{Width: 336, Height: 106}
	other := Rect{Left: 900, Top: 106, Right: 1076, Bottom: 302}
	withPullOut := func(at Point) []Rect {
		ribbon := rectOf(at, size)
		pullOut := Rect{Left: at.X + (336-480)/2, Top: ribbon.Bottom, Right: at.X + (336-480)/2 + 480, Bottom: ribbon.Bottom + 240}
		return []Rect{ribbon, pullOut}
	}
	got, ok := Clear(Point{X: 792, Y: 0}, size, work, false, []Rect{other}, withPullOut)
	if !ok {
		t.Fatal("no clear place")
	}
	for _, part := range withPullOut(got) {
		if overlap(part, other) > 0 {
			t.Errorf("at %+v its footprint %+v still covers %+v", got, part, other)
		}
	}
	if got != (Point{X: 492, Y: 0}) {
		t.Errorf("got %+v, want (492, 0): its pull out ending where the other begins", got)
	}
}

// FR-412: the edge across from each edge, where a ribbon goes when its own edge has no clear place.
func TestEachEdgeHasAnOpposite(t *testing.T) {
	t.Parallel()
	for edge, want := range map[Edge]Edge{Left: Right, Right: Left, Top: Bottom, Bottom: Top} {
		if got := Opposite(edge); got != want {
			t.Errorf("%s: got %s, want %s", edge, got, want)
		}
	}
	if got := Opposite(""); got != "" {
		t.Errorf("no edge: got %q", got)
	}
}
