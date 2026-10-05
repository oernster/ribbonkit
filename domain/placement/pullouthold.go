package placement

// PullOutHeld answers the pull out while the ribbon's grip is dragged (FR-623): the pull out as it
// stood when the drag began (held, beside heldRibbon), its size and its place along the ribbon
// measured from the ribbon's top-left corner kept, still adjoining side of ribbon as the ribbon now
// is. A pull out that followed the ribbon's size moved the window's top-left corner at every step of
// the drag, which shifted the clocks drawn from that corner (measured 2026-10-04); held, the corner
// stays put.
func PullOutHeld(ribbon Rect, side Edge, held, heldRibbon Rect) Rect {
	size := Size{Width: held.Width(), Height: held.Height()}
	along := Point{X: held.Left - heldRibbon.Left, Y: held.Top - heldRibbon.Top}
	var at Point
	switch side {
	case Bottom:
		at = Point{X: ribbon.Left + along.X, Y: ribbon.Bottom}
	case Top:
		at = Point{X: ribbon.Left + along.X, Y: ribbon.Top - size.Height}
	case Right:
		at = Point{X: ribbon.Right, Y: ribbon.Top + along.Y}
	default:
		at = Point{X: ribbon.Left - size.Width, Y: ribbon.Top + along.Y}
	}
	return rectOf(at, size)
}
