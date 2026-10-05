package ribbon

// ScaleAfter answers the scale a drag of the corner grip has reached (FR-623): the scale it began
// at, grown or shrunk as the ribbon's thickness would be by moving its far side the distance moved,
// held within MinScale to MaxScale. It is not rounded, so the ribbon follows the pointer pixel by
// pixel while the grip is dragged; the scale kept is rounded to a whole percent. Moved and
// thickness are in the same units; moved is positive outward. A thickness that is not positive
// gives the scale the drag began at.
func ScaleAfter(began, thickness, moved float64) float64 {
	if !(thickness > 0) {
		return began
	}
	scale := began * (thickness + moved) / thickness
	return min(max(scale, MinScale), MaxScale)
}
