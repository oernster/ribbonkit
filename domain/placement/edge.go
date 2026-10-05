package placement

// SnapReach is how near, in DIP, a dropped ribbon's side must come to an edge to be put flush
// against it (FR-410).
const SnapReach = 16

// Against is an edge of one display's work area: the edge a ribbon last stood flush against
// (FR-411).
type Against struct {
	// Device is the display's device name, such as `\\.\DISPLAY2`.
	Device string
	Edge   Edge
}

// Along answers whether edge runs along a ribbon of the orientation vertical tells: the left and
// right edges for a vertical ribbon, the top and bottom for a horizontal one (FR-410, FR-619).
func Along(edge Edge, vertical bool) bool {
	if vertical {
		return edge == Left || edge == Right
	}
	return edge == Top || edge == Bottom
}

// FlushAgainst answers the edge of work a ribbon at at of size lies flush against, wholly inside
// work, among the edges running along its orientation (FR-619); false when it lies against none. A
// ribbon as broad as work touches both; it is the far one, right or bottom, that counts.
func FlushAgainst(at Point, size Size, work Rect, vertical bool) (Edge, bool) {
	ribbon := rectOf(at, size)
	if ribbon.Left < work.Left || ribbon.Right > work.Right || ribbon.Top < work.Top || ribbon.Bottom > work.Bottom {
		return "", false
	}
	near, far := Left, Right
	nearFlush, farFlush := ribbon.Left == work.Left, ribbon.Right == work.Right
	if !vertical {
		near, far = Top, Bottom
		nearFlush, farFlush = ribbon.Top == work.Top, ribbon.Bottom == work.Bottom
	}
	switch {
	case farFlush:
		return far, true
	case nearFlush:
		return near, true
	}
	return "", false
}

// Snapped answers at with a ribbon of size put flush against the edge of work running along its
// orientation that its side lies within reach of, on either side of that edge, its position along
// the edge kept; the nearer edge where both are within reach; at itself where neither is (FR-410).
// Then clamped into work.
func Snapped(at Point, size Size, work Rect, vertical bool, reach int) Point {
	position, length, low, high := at.X, size.Width, work.Left, work.Right
	if !vertical {
		position, length, low, high = at.Y, size.Height, work.Top, work.Bottom
	}
	toLow, toHigh := abs(position-low), abs(high-(position+length))
	switch {
	case toLow <= reach && toLow <= toHigh:
		position = low
	case toHigh <= reach:
		position = high - length
	}
	if vertical {
		at.X = position
	} else {
		at.Y = position
	}
	return Clamp(at, size, work)
}

func rectOf(at Point, size Size) Rect {
	return Rect{Left: at.X, Top: at.Y, Right: at.X + size.Width, Bottom: at.Y + size.Height}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
