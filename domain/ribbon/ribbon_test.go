package ribbon

import (
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// FR-103, FR-505, FR-613: a first run is Classic, vertical, the system's theme, pinned, not on top,
// unplaced and with no edge remembered.
func TestDefaultsAreVerticalPinnedAndNotOnTop(t *testing.T) {
	t.Parallel()
	got := Defaults()
	if got.Colour != Classic || got.Orientation != Vertical || got.Theme != System || got.AlwaysOnTop || got.Placement != nil {
		t.Errorf("got %+v", got)
	}
	if !got.Pinned || got.OnTop(true) || got.LastEdge != nil {
		t.Errorf("a first run is pinned %v, on top %v, edge %v; want pinned, not on top, no edge", got.Pinned, got.OnTop(true), got.LastEdge)
	}
}

// FR-622: a first run is wholly opaque; an opacity outside the bounds, a hand-edited one or none at
// all, is brought to the nearer bound; one inside them is kept.
func TestOpacityIsHeldWithinItsBounds(t *testing.T) {
	t.Parallel()
	if got := Defaults().Opacity; got != MaxOpacity {
		t.Errorf("a first run is %d percent opaque", got)
	}
	for stored, want := range map[int]int{0: MinOpacity, MinOpacity - 1: MinOpacity, MinOpacity: MinOpacity, 55: 55, MaxOpacity: MaxOpacity, MaxOpacity + 1: MaxOpacity} {
		if got := (Choices{Opacity: stored}).Normalised().Opacity; got != want {
			t.Errorf("%d normalised to %d, want %d", stored, got, want)
		}
	}
}

// FR-623: a first run draws cells as they are; a scale outside the bounds is brought to the nearer.
func TestScaleIsHeldWithinItsBounds(t *testing.T) {
	t.Parallel()
	if got := Defaults().Scale; got != WholeScale {
		t.Errorf("a first run is scaled %d percent", got)
	}
	for stored, want := range map[int]int{0: MinScale, MinScale - 1: MinScale, 150: 150, MaxScale + 1: MaxScale} {
		if got := (Choices{Scale: stored}).Normalised().Scale; got != want {
			t.Errorf("%d normalised to %d, want %d", stored, got, want)
		}
	}
}

// FR-617, FR-619: a ribbon unpinned in effect (unpinned and flush) stays on top whatever Always on
// top holds; for one pinned (or unpinned away from every edge) Always on top decides.
func TestAnUnpinnedRibbonIsAlwaysOnTop(t *testing.T) {
	t.Parallel()
	for _, each := range []struct{ alwaysOnTop, pinned, flush, want bool }{
		{false, true, true, false}, {true, true, true, true}, {false, false, true, true}, {true, false, true, true},
		{false, false, false, false}, {true, false, false, true},
	} {
		c := Defaults()
		c.AlwaysOnTop, c.Pinned = each.alwaysOnTop, each.pinned
		if got := c.OnTop(each.flush); got != each.want {
			t.Errorf("Always on top %v, pinned %v, flush %v: on top %v, want %v", each.alwaysOnTop, each.pinned, each.flush, got, each.want)
		}
	}
}

// FR-619: the pin in effect; the choice itself is never changed by it.
func TestFlushnessGivesThePinInEffect(t *testing.T) {
	t.Parallel()
	for _, each := range []struct{ pinned, flush, want bool }{
		{true, true, true}, {true, false, true}, {false, true, false}, {false, false, true},
	} {
		c := Defaults()
		c.Pinned = each.pinned
		if got := c.PinnedInEffect(each.flush); got != each.want || c.Pinned != each.pinned {
			t.Errorf("pinned %v, flush %v: in effect %v, want %v", each.pinned, each.flush, got, each.want)
		}
	}
}

// FR-411: a remembered edge naming no edge, as a hand edit might, is forgotten; a real one is kept.
func TestAnUnknownRememberedEdgeIsForgotten(t *testing.T) {
	t.Parallel()
	c := Defaults()
	c.LastEdge = &placement.Against{Device: `\\.\DISPLAY1`, Edge: "middle"}
	if got := c.Normalised(); got.LastEdge != nil {
		t.Errorf("kept %+v", got.LastEdge)
	}
	c.LastEdge = &placement.Against{Device: `\\.\DISPLAY1`, Edge: placement.Left}
	if got := c.Normalised(); got.LastEdge == nil || *got.LastEdge != *c.LastEdge {
		t.Errorf("lost %+v", c.LastEdge)
	}
}

// FR-902, Amendment 35: a kept side naming no side, as a hand edit might, is forgotten; a real one is
// kept; a first run has none.
func TestAnUnknownPullOutSideIsForgotten(t *testing.T) {
	t.Parallel()
	c := Defaults()
	if c.PullOutSide != "" {
		t.Errorf("a first run keeps the side %q", c.PullOutSide)
	}
	c.PullOutSide = "middle"
	if got := c.Normalised(); got.PullOutSide != "" {
		t.Errorf("kept %q", got.PullOutSide)
	}
	c.PullOutSide = placement.Left
	if got := c.Normalised(); got.PullOutSide != placement.Left {
		t.Errorf("lost the left side, got %q", got.PullOutSide)
	}
}

// FR-409: a horizontal ribbon goes to the top edge, a vertical one to the right.
func TestEachOrientationHasAHomeEdge(t *testing.T) {
	t.Parallel()
	for orientation, want := range map[Orientation]placement.Edge{Horizontal: placement.Top, Vertical: placement.Right} {
		if got, ok := HomeEdge(orientation); !ok || got != want {
			t.Errorf("%s: got %s, %v; want %s", orientation, got, ok, want)
		}
	}
	if _, ok := HomeEdge("diagonal"); ok {
		t.Error("an orientation the setting does not offer has a home edge")
	}
}

// A choice holding a word it does not offer, as a hand edit might, is its default; a known one is
// kept; a choice with no fixed set of words is left alone.
func TestUnknownChoicesAreNormalisedToDefaults(t *testing.T) {
	t.Parallel()
	got := (Choices{Colour: "mauve", Orientation: "diagonal", Theme: "sepia", AlwaysOnTop: true}).Normalised()
	want := Defaults()
	if got.Colour != want.Colour || got.Orientation != want.Orientation || got.Theme != want.Theme || !got.AlwaysOnTop {
		t.Errorf("got %+v", got)
	}
	for _, known := range []Choices{
		{Colour: Neon, Orientation: Horizontal, Theme: Dark},
		{Colour: Contrast, Orientation: Vertical, Theme: Light},
		{Colour: Classic, Orientation: Vertical, Theme: System},
	} {
		if kept := known.Normalised(); kept.Colour != known.Colour || kept.Orientation != known.Orientation || kept.Theme != known.Theme {
			t.Errorf("known choices %+v were changed to %+v", known, kept)
		}
	}
}
