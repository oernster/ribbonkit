package ribbon

import "strings"

// MaxLabelLength is the most characters a cell's label holds.
const MaxLabelLength = 32

// Label answers the label to store for what the user typed: trimmed of surrounding spaces and cut to
// MaxLabelLength characters; fallback, the product's own default for the cell, when nothing is left.
func Label(typed, fallback string) string {
	trimmed := strings.TrimSpace(typed)
	if trimmed == "" {
		return fallback
	}
	characters := []rune(trimmed)
	if len(characters) > MaxLabelLength {
		return strings.TrimSpace(string(characters[:MaxLabelLength]))
	}
	return trimmed
}
