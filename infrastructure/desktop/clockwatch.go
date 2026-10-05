//go:build !windows

package desktop

import "time"

// Off Windows no broadcast says the time was set or the machine woke, so both are seen the same
// way: the wall clock moving by a different amount from the monotonic clock between two checks.
// Setting the time moves only the wall clock; the monotonic clock does not advance while the
// machine sleeps, so waking shows the wall clock ahead of it.
const (
	clockCheckEvery = 2 * time.Second
	clockTolerance  = 2 * time.Second
)

// jumped answers whether the wall clock moved by more than clockTolerance apart from the monotonic
// clock over one interval.
func jumped(wall, monotonic time.Duration) bool {
	drift := wall - monotonic
	if drift < 0 {
		drift = -drift
	}
	return drift > clockTolerance
}

// watchClock calls report whenever the wall clock jumps, until stop is closed.
func watchClock(stop <-chan struct{}, report func()) {
	ticker := time.NewTicker(clockCheckEvery)
	defer ticker.Stop()
	last := time.Now()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			now := time.Now()
			if jumped(now.Round(0).Sub(last.Round(0)), now.Sub(last)) {
				report()
			}
			last = now
		}
	}
}
