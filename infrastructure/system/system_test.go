package system

import (
	"testing"
	"time"
)

func TestTheClockReadsTheWindowsClock(t *testing.T) {
	t.Parallel()
	before := time.Now()
	got := Clock{}.Now()
	if got.Before(before) || got.Sub(before) > time.Second {
		t.Errorf("got %s, read %s", got, before)
	}
}

func TestIdsAreDistinct(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for range 1000 {
		id := IDs{}.NewID()
		if id == "" || seen[id] {
			t.Fatalf("id %q repeated or empty", id)
		}
		seen[id] = true
	}
}
