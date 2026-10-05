//go:build windows

package setup

import (
	"strconv"
	"strings"
)

// Relation says how the version setup carries stands against the one installed. It decides
// whether setup offers an update, the Installed screen or a way back (FR-801).
type Relation int

const (
	// Older means setup carries an earlier version than the one installed.
	Older Relation = iota - 1
	// Same means the two versions match.
	Same
	// Newer means setup carries a later version.
	Newer
)

// fields is the count of numeric fields compared: major, minor and patch.
const fields = 3

// suffixMark begins a pre-release suffix, which takes no part in the comparison.
const suffixMark = "-"

// Compare answers how version a stands against version b: major, minor then patch as numbers,
// ignoring anything after a hyphen, with a missing or non-numeric field counting as zero (FR-801).
func Compare(a, b string) Relation {
	left, right := parse(a), parse(b)
	for i := range fields {
		switch {
		case left[i] > right[i]:
			return Newer
		case left[i] < right[i]:
			return Older
		}
	}
	return Same
}

// parse reads the leading major.minor.patch of a version.
func parse(version string) [fields]int {
	core, _, _ := strings.Cut(strings.TrimSpace(version), suffixMark)
	parts := strings.Split(core, ".")
	var out [fields]int
	for i := 0; i < fields && i < len(parts); i++ {
		if number, err := strconv.Atoi(strings.TrimSpace(parts[i])); err == nil {
			out[i] = number
		}
	}
	return out
}
