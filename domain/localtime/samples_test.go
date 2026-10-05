package localtime

import (
	"slices"
	"testing"
	"time"
)

// Every minute of the day once, in the chosen time format.
func TestSamplesHoldEveryTimeOnce(t *testing.T) {
	t.Parallel()
	minutes := int(Day / time.Minute)
	for format, want := range map[Format][]string{
		TwentyFourHour: {"00:00", "09:05", "12:30", "23:59"},
		TwelveHour:     {"12:00 AM", "9:05 AM", "12:30 PM", "11:59 PM"},
	} {
		times := TimeSamples(format)
		if len(times) != minutes {
			t.Errorf("%s: %d times, want %d", format, len(times), minutes)
		}
		for _, each := range want {
			if !slices.Contains(times, each) {
				t.Errorf("%s: %q missing", format, each)
			}
		}
	}
}

// Each text is answered once, in the order first met.
func TestDistinctKeepsTheFirstOfEach(t *testing.T) {
	t.Parallel()
	var start time.Time
	got := Distinct(start, start.Add(4*time.Hour), func(at time.Time) time.Time { return at.Add(time.Hour) },
		func(at time.Time) string { return []string{"a", "b", "a", "c"}[at.Hour()] })
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("answered %v", got)
	}
}
