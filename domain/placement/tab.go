package placement

// TabThickness is how deep, in DIP, the tab of an unpinned ribbon is (FR-614).
const TabThickness = 8

// Tab answers the rectangle a ribbon at at of size, flush against edge, shrinks to (FR-614): a
// band thickness pixels deep along the ribbon's whole length, covering the side against edge. Only
// a flush ribbon collapses (FR-619), so the band always lies on its edge.
func Tab(at Point, size Size, edge Edge, thickness int) Rect {
	ribbon := rectOf(at, size)
	switch edge {
	case Left:
		ribbon.Right = ribbon.Left + thickness
	case Right:
		ribbon.Left = ribbon.Right - thickness
	case Top:
		ribbon.Bottom = ribbon.Top + thickness
	case Bottom:
		ribbon.Top = ribbon.Bottom - thickness
	}
	return ribbon
}
