package placement

// Keeping two ribbons apart (FR-412). Each running ribbon holds the rectangles it occupies; the one
// being placed slides along its own length to the nearest place where nothing of its own lands on
// them. Only the ribbon being placed moves.

// Clear answers where a ribbon wanted at at of size stands so that its footprint overlaps none of
// taken (FR-412): at itself where it already does; else the nearest place along its own length
// (along y for a vertical ribbon, x for a horizontal one) inside work, its position across kept.
// footprint answers the rectangles the ribbon occupies standing at a point: the ribbon and its pull
// out as it would show there. Of two places equally near, the one towards the top or left wins.
// Touching is not overlapping. It answers false where no place on that line is clear.
func Clear(at Point, size Size, work Rect, vertical bool, taken []Rect, footprint func(Point) []Rect) (Point, bool) {
	clear := func(candidate Point) bool {
		if Clamp(candidate, size, work) != candidate {
			return false
		}
		for _, part := range footprint(candidate) {
			for _, other := range taken {
				if overlap(part, other) > 0 {
					return false
				}
			}
		}
		return true
	}
	if clear(at) {
		return at, true
	}
	along := func(p Point) int { return p.X }
	moved := func(position int) Point { return Point{X: position, Y: at.Y} }
	if vertical {
		along = func(p Point) int { return p.Y }
		moved = func(position int) Point { return Point{X: at.X, Y: position} }
	}
	wanted := along(at)
	best, found := 0, false
	for _, position := range candidates(at, size, work, vertical, taken, footprint) {
		candidate := moved(position)
		if !clear(candidate) {
			continue
		}
		distance, bestDistance := abs(position-wanted), abs(best-wanted)
		if !found || distance < bestDistance || (distance == bestDistance && position < best) {
			best, found = position, true
		}
	}
	return moved(best), found
}

// candidates answers the positions along the ribbon's length worth trying: each part of its
// footprint, as it lies at at, brought to touch each taken rectangle from either side, plus the two
// ends of work. Clear keeps those that are clear; where a pull out shifts as the ribbon moves, a
// position may not touch exactly, which only makes it a little farther than it need be.
func candidates(at Point, size Size, work Rect, vertical bool, taken []Rect, footprint func(Point) []Rect) []int {
	start := func(r Rect) int { return r.Left }
	end := func(r Rect) int { return r.Right }
	origin, length, low, high := at.X, size.Width, work.Left, work.Right
	if vertical {
		start = func(r Rect) int { return r.Top }
		end = func(r Rect) int { return r.Bottom }
		origin, length, low, high = at.Y, size.Height, work.Top, work.Bottom
	}
	positions := []int{low, high - length}
	for _, part := range footprint(at) {
		before, after := start(part)-origin, end(part)-origin
		for _, other := range taken {
			positions = append(positions, start(other)-after, end(other)-before)
		}
	}
	return positions
}

// Opposite answers the edge across work from edge (FR-412, FR-902): left for right, top for bottom
// and each the other way; no edge for none.
func Opposite(edge Edge) Edge {
	switch edge {
	case Top:
		return Bottom
	case Bottom:
		return Top
	case Left:
		return Right
	case Right:
		return Left
	}
	return ""
}
