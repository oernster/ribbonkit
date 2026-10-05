package ribbon

import (
	"math"
	"testing"
)

// FR-623: the far side follows the pointer, so moving it a quarter of the thickness outward grows
// the scale by a quarter; inward shrinks it; a pixel moves it by a fraction; the bounds hold; no
// thickness keeps the scale.
func TestScaleAfterFollowsTheFarSide(t *testing.T) {
	t.Parallel()
	const thickness = 200.0
	cases := []struct {
		name  string
		moved float64
		want  float64
	}{
		{"still", 0, WholeScale},
		{"a quarter outward", thickness / 4, WholeScale + WholeScale/4},
		{"a fifth inward", -thickness / 5, WholeScale - WholeScale/5},
		{"one unit outward", 1, WholeScale + WholeScale/thickness},
		{"far outward", thickness * 4, MaxScale},
		{"far inward", -thickness, MinScale},
	}
	for _, each := range cases {
		if got := ScaleAfter(WholeScale, thickness, each.moved); got != each.want {
			t.Errorf("%s: %v, want %v", each.name, got, each.want)
		}
	}
	for _, unusable := range []float64{0, -thickness, math.NaN()} {
		if got := ScaleAfter(WholeScale+1, unusable, thickness); got != WholeScale+1 {
			t.Errorf("thickness %v: %v, want the scale it began at", unusable, got)
		}
	}
}
