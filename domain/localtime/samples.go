package localtime

import "time"

// Day is how long a day of local time runs, every minute of which a cell may show.
const Day = 24 * time.Hour

// TimeSamples answers every distinct time a cell can show in format, every minute of a day once. They
// are what the page measures, in the font it really draws with, to find how wide a cell must be for
// its widest time.
func TimeSamples(format Format) []string {
	var start time.Time
	return Distinct(start, start.Add(Day), func(at time.Time) time.Time { return at.Add(time.Minute) },
		func(at time.Time) string { return Text(at, format) })
}

// Distinct answers each instant from start up to end, stepped by next, as write writes it, once each
// in the order first met.
func Distinct(start, end time.Time, next func(time.Time) time.Time, write func(time.Time) string) []string {
	seen := map[string]bool{}
	written := []string{}
	for at := start; at.Before(end); at = next(at) {
		text := write(at)
		if !seen[text] {
			seen[text] = true
			written = append(written, text)
		}
	}
	return written
}
