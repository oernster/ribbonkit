package arranger

import (
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// Neighbours is what the arranger knows of the other ribbons running for the same user, of any
// product built on the kit, so that none lands on another (FR-412). Neither call fails: an
// implementation that cannot read or write says so in its own log and answers as though alone.
type Neighbours interface {
	// Taken answers the rectangles the other running ribbons occupy, in physical pixels: each
	// ribbon in full and its pull out while shown.
	Taken() []placement.Rect
	// Hold records the rectangles this ribbon occupies now, for the others to keep off.
	Hold(occupied []placement.Rect)
}

// NoNeighbours is a ribbon alone on its desktop: nothing is taken and nothing is held.
type NoNeighbours struct{}

// Taken answers that no other ribbon is running.
func (NoNeighbours) Taken() []placement.Rect { return nil }

// Hold keeps nothing.
func (NoNeighbours) Hold([]placement.Rect) {}

// clearOfOthers answers placed's corner moved so that the ribbon, with its pull out as it would show
// there, overlaps no other ribbon (FR-412): the nearest clear place along the edge it stands against,
// else along the opposite edge; where neither has one, where it was. Nothing moves while the grip is
// dragged; the ribbon is cleared once the drag ends.
func (a *Arranger) clearOfOthers(current ribbon.Choices, content Content, placed placement.Placed, size placement.Size) placement.Point {
	if a.gripDragged() {
		return placed.At
	}
	taken := a.neighbours.Taken()
	if len(taken) == 0 {
		return placed.At
	}
	vertical := current.Orientation == ribbon.Vertical
	work := placed.Monitor.Work
	clearFrom := func(at placement.Point) (placement.Point, bool) {
		edge, _ := placement.FlushAgainst(at, size, work, vertical)
		footprint := func(corner placement.Point) []placement.Rect {
			_, pullOut, shown := a.pullOutBeside(current, content, corner, size, placed.Monitor, edge)
			return footprintOf(corner, size, pullOut, shown)
		}
		return placement.Clear(at, size, work, vertical, taken, footprint)
	}
	if at, ok := clearFrom(placed.At); ok {
		return at
	}
	if edge, flush := placement.FlushAgainst(placed.At, size, work, vertical); flush {
		if at, ok := clearFrom(placement.AgainstEdge(size, work, placement.Opposite(edge))); ok {
			return at
		}
	}
	return placed.At
}

// footprintOf answers what a ribbon of size at at occupies: itself, with its pull out while shown.
func footprintOf(at placement.Point, size placement.Size, pullOut placement.Rect, shown bool) []placement.Rect {
	parts := []placement.Rect{{Left: at.X, Top: at.Y, Right: at.X + size.Width, Bottom: at.Y + size.Height}}
	if shown {
		parts = append(parts, pullOut)
	}
	return parts
}

// gripDragged answers whether the ribbon's grip is being dragged now (FR-623).
func (a *Arranger) gripDragged() bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.previewScale != 0
}
