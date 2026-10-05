package ribbon

import (
	"strings"
	"testing"
)

// A label empty after trimming is the product's fallback; a typed one is kept trimmed.
func TestEmptyLabelFallsBackToTheProductsDefault(t *testing.T) {
	t.Parallel()
	for _, typed := range []string{"", "   ", "\t"} {
		if got := Label(typed, "London"); got != "London" {
			t.Errorf("Label(%q) = %q, want the fallback London", typed, got)
		}
	}
	if got := Label("  Brighton ", "London"); got != "Brighton" {
		t.Errorf("a typed label is kept trimmed: got %q", got)
	}
}

// A label is cut to 32 characters, counted as characters rather than bytes, with no trailing space.
func TestLabelIsCappedAt32Characters(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("é", MaxLabelLength+5)
	if got := []rune(Label(long, "Paris")); len(got) != MaxLabelLength {
		t.Errorf("got %d characters, want %d", len(got), MaxLabelLength)
	}
	exact := strings.Repeat("a", MaxLabelLength)
	if got := Label(exact, "Paris"); got != exact {
		t.Errorf("a label of exactly the limit is kept whole: got %q", got)
	}
	spaced := strings.Repeat("a", MaxLabelLength-1) + " b"
	if got := Label(spaced, "Paris"); got != strings.Repeat("a", MaxLabelLength-1) {
		t.Errorf("a cut ending in a space is trimmed: got %q", got)
	}
}
