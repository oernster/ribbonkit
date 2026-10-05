package placement

// Fitted is the ribbon's length along its orientation and whether its cells scroll.
type Fitted struct {
	Length  int
	Scrolls bool
}

// Fit answers the ribbon's length along its orientation (FR-105, FR-106): the cells' lengths plus
// padding at each end, while that fits available; else available itself with the cells scrolling.
// No clocks still leaves room for one cell, which holds the empty ribbon's prompt (FR-107).
func Fit(cells, cellLength, padding, available int) Fitted {
	wanted := max(cells, 1)*cellLength + 2*padding
	if wanted > available {
		return Fitted{Length: available, Scrolls: true}
	}
	return Fitted{Length: wanted}
}
