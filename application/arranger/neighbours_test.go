package arranger

import (
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// otherRibbon is FR-412's worked example: another product's vertical ribbon, 176 by 196, against the
// primary's right edge, centred, from (1744, 418) to (1920, 614).
var otherRibbon = placement.Rect{Left: 1744, Top: 418, Right: 1920, Bottom: 614}

// overlaps answers whether any part of footprint overlaps other, touching not counting.
func overlaps(footprint []placement.Rect, other placement.Rect) bool {
	return slices.ContainsFunc(footprint, func(part placement.Rect) bool {
		return min(part.Right, other.Right) > max(part.Left, other.Left) && min(part.Bottom, other.Bottom) > max(part.Top, other.Top)
	})
}

// FR-412: a ribbon launched with nothing stored, 286 tall, would centre from 373 to 659 over the other
// ribbon; it ties at 241 between above and below and stands above, still against the right edge.
func TestALaunchedRibbonKeepsOffAnother(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(3))
	r.neighbours.occupy(otherRibbon)
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1744, Y: 132}) || got.Edge != placement.Right {
		t.Errorf("launched at %+v against %q, want (1744, 132) against the right edge", got.At, got.Edge)
	}
}

// FR-412: a ribbon dropped against the right edge with its top at 500 stands below the other (114
// away rather than 368 above); that place is stored.
func TestADroppedRibbonKeepsOffAnotherAndIsStoredThere(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(3))
	r.neighbours.occupy(otherRibbon)
	got, err := r.arranger.Moved(placement.Point{X: 1744, Y: 500})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1744, Y: 614}) {
		t.Errorf("dropped at %+v, want (1744, 614)", got.At)
	}
	if stored := r.host.last(t).Placement; stored == nil || stored.Offset != got.At {
		t.Errorf("stored %+v, want the cleared place", stored)
	}
}

// FR-412: a ribbon 916 tall finds no clear place on the right edge, 418 above and 418 below, so it
// stands against the left edge, centred.
func TestWithNoRoomOnItsEdgeARibbonTakesTheOpposite(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(10))
	r.neighbours.occupy(otherRibbon)
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 0, Y: 58}) || got.Edge != placement.Left {
		t.Errorf("launched at %+v against %q, want (0, 58) against the left edge", got.At, got.Edge)
	}
}

// FR-412: where neither its edge nor the opposite one has a clear place, the ribbon stands where it
// was asked to, since nowhere is better.
func TestWithNoRoomOnEitherEdgeARibbonStaysWhereItWasPut(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(10))
	r.neighbours.occupy(otherRibbon, placement.Rect{Left: 0, Top: 418, Right: 176, Bottom: 614})
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1744, Y: 58}) {
		t.Errorf("launched at %+v, want the default place (1744, 58)", got.At)
	}
}

// FR-412, OQ-38: a horizontal ribbon at the top whose pull out, 480 by 240 below it, would cover
// another ribbon its own body clears moves along until the pull out clears too: to x 492, its pull
// out ending at 900 where the other begins.
func TestTheRibbonsOwnPullOutKeepsOffAnother(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), pulledOut(2))
	other := placement.Rect{Left: 900, Top: 118, Right: 1076, Bottom: 314}
	r.neighbours.occupy(other)
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 492, Y: 0}) || overlaps([]placement.Rect{got.PullOut}, other) {
		t.Errorf("launched at %+v with its pull out at %+v, want (492, 0) clear of %+v", got.At, got.PullOut, other)
	}
}

// FR-412: nothing is cleared while the grip is dragged, only once the scale is kept. Growing at the
// right edge the ribbon is clamped back inside it as ever; it is not slid along its length off the
// other while the drag lasts.
func TestARibbonIsClearedOnlyOnceTheGripIsLetGo(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(2))
	before, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	other := placement.Rect{Left: 1744, Top: 500, Right: 1920, Bottom: 700}
	r.neighbours.occupy(other)
	if err := r.arranger.PreviewScale(ribbon.WholeScale + ribbon.WholeScale/2); err != nil {
		t.Fatal(err)
	}
	during, err := r.arranger.Rearrange(before.At)
	if err != nil || during.At.Y != before.At.Y || !overlaps(r.neighbours.lastHeld(t), other) {
		t.Fatalf("during the drag it stood at %+v (%v); want its top kept at %d over the other", during.At, err, before.At.Y)
	}
	if err := r.arranger.SetScale(ribbon.WholeScale + ribbon.WholeScale/2); err != nil {
		t.Fatal(err)
	}
	kept, err := r.arranger.Rearrange(during.At)
	if err != nil {
		t.Fatal(err)
	}
	if overlaps(r.neighbours.lastHeld(t), other) {
		t.Errorf("once let go it still stands over the other at %+v", kept.At)
	}
}

// FR-412: the ribbon holds what it occupies once arranged: itself, with its pull out while shown.
func TestTheRibbonHoldsWhatItOccupies(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	ribbonRect := placement.Rect{Left: got.At.X, Top: got.At.Y, Right: got.At.X + got.Size.Width, Bottom: got.At.Y + got.Size.Height}
	if held := r.neighbours.lastHeld(t); !slices.Equal(held, []placement.Rect{ribbonRect}) {
		t.Errorf("held %+v, want the ribbon alone %+v", held, ribbonRect)
	}
	r = newRig(horizontal(), pulledOut(2))
	got, err = r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if held := r.neighbours.lastHeld(t); len(held) != 2 || held[1] != got.PullOut {
		t.Errorf("held %+v, want the ribbon and its pull out %+v", held, got.PullOut)
	}
}

// FR-408, FR-412: launched beside the other ribbon at (1744, 132), centring on the right edge would
// clear it back to the same place, so that press would not move it; the left edge would. Asking saves
// nothing and holds nothing.
func TestAnEdgeWhoseCentreIsTakenDoesNotMoveARibbonAlreadyBesideIt(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(3))
	r.neighbours.occupy(otherRibbon)
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	saves, held := r.host.saves(), len(r.neighbours.held)
	if r.arranger.EdgeMoves(placement.Right) || !r.arranger.EdgeMoves(placement.Left) {
		t.Errorf("right moves %v, left %v; want false then true", r.arranger.EdgeMoves(placement.Right), r.arranger.EdgeMoves(placement.Left))
	}
	if r.host.saves() != saves || len(r.neighbours.held) != held {
		t.Errorf("asking saved %d and held %d more, want none", r.host.saves()-saves, len(r.neighbours.held)-held)
	}
}

// FR-408: a ribbon alone already centred on its edge would not move; dropped lower on that edge, it
// would, back to the centre.
func TestAnEdgeMovesARibbonOnlyWhereItStandsElsewhere(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(2))
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	if r.arranger.EdgeMoves(placement.Right) {
		t.Error("a ribbon centred on the right edge would move there, want not")
	}
	if _, err := r.arranger.Moved(placement.Point{X: 1744, Y: 600}); err != nil {
		t.Fatal(err)
	}
	if !r.arranger.EdgeMoves(placement.Right) {
		t.Error("a ribbon lower on the right edge would not move to its centre, want it to")
	}
}

// FR-408: every edge is offered before the ribbon has been arranged and where the displays cannot be
// read, so pressing it reports what went wrong.
func TestAnEdgeIsOfferedWhereItCannotBeKnownWhetherItMoves(t *testing.T) {
	t.Parallel()
	r := newRig(vertical(), cells(2))
	if !r.arranger.EdgeMoves(placement.Right) {
		t.Error("an edge was withheld before the ribbon was arranged")
	}
	if _, err := r.arranger.Launch(); err != nil {
		t.Fatal(err)
	}
	r.monitors.err = errPlanted
	if !r.arranger.EdgeMoves(placement.Right) {
		t.Error("an edge was withheld while the displays could not be read")
	}
}

// FR-412: no neighbours given is a ribbon alone, placed where it would be with none running.
func TestNoNeighboursIsARibbonAlone(t *testing.T) {
	t.Parallel()
	host := &fakeHost{choices: vertical(), content: cells(2)}
	alone := New(host, &fakeMonitors{monitors: []placement.Monitor{primaryMonitor}}, nil)
	got, err := alone.Launch()
	if err != nil || got.At != (placement.Point{X: 1744, Y: 418}) {
		t.Errorf("launched at %+v (%v), want the default place (1744, 418)", got.At, err)
	}
}
