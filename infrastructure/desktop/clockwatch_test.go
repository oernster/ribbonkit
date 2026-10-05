//go:build !windows

package desktop

import (
	"testing"
	"time"
)

// FR-209: a wall clock that keeps pace is quiet. A jump is one set forward or back; so is one that
// ran on while the machine slept.
func TestAWallClockJumpIsSeenAndSteadyTimeIsNot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name            string
		wall, monotonic time.Duration
		want            bool
	}{
		{"steady", clockCheckEvery, clockCheckEvery, false},
		{"within tolerance", clockCheckEvery + clockTolerance, clockCheckEvery, false},
		{"set forward", time.Hour, clockCheckEvery, true},
		{"set back", -time.Hour, clockCheckEvery, true},
		{"slept", 8 * time.Hour, clockCheckEvery, true},
	}
	for _, c := range cases {
		if got := jumped(c.wall, c.monotonic); got != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
}

// watchClock stops when told to.
func TestTheClockWatchStops(t *testing.T) {
	t.Parallel()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		watchClock(stop, func() {})
		close(done)
	}()
	close(stop)
	<-done
}
