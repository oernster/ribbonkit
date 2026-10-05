package desktop

import (
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// A display 1920 by 1032 with a second to its right, as moves across them are proposed.
var workLeft = placement.Rect{Right: 1920, Bottom: 1032}

func TestAMovePastAnEdgeStopsAtIt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		proposed rect
		want     rect
	}{
		{"inside stays", rect{100, 100, 288, 204}, rect{100, 100, 288, 204}},
		{"past the right edge", rect{1800, 100, 1988, 204}, rect{1732, 100, 1920, 204}},
		{"past the top", rect{100, -40, 288, 64}, rect{100, 0, 288, 104}},
		{"past the bottom left corner", rect{-50, 1000, 138, 1104}, rect{0, 928, 188, 1032}},
	}
	for _, each := range cases {
		if got := fenced(each.proposed, workLeft); got != each.want {
			t.Errorf("%s: got %+v, want %+v", each.name, got, each.want)
		}
	}
}

// On another display the same move is fenced by that display's work area instead.
func TestAMoveOntoAnotherDisplayIsFencedByIt(t *testing.T) {
	t.Parallel()
	workRight := placement.Rect{Left: 1920, Right: 4480, Bottom: 1392}
	got := fenced(rect{1900, 1300, 2088, 1404}, workRight)
	if got != (rect{1920, 1288, 2108, 1392}) {
		t.Errorf("got %+v", got)
	}
}

// The display under a point is read from Windows; the primary holds (0, 0) on every machine.
func TestTheWorkAreaUnderAPointIsRead(t *testing.T) {
	t.Parallel()
	work, ok := workAreaAt(point{x: 1, y: 1})
	if !ok || work.Width() <= 0 || work.Height() <= 0 {
		t.Errorf("got %+v %v", work, ok)
	}
}
