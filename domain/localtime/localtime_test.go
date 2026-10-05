package localtime

import (
	"slices"
	"testing"
	"time"
)

// instant parses an RFC 3339 instant or fails the test.
func instant(t *testing.T, text string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		t.Fatalf("parsing %s: %v", text, err)
	}
	return parsed
}

func TestTwelveAndTwentyFourHourFormats(t *testing.T) {
	t.Parallel()
	cases := []struct{ at, twentyFour, twelve string }{
		{"2026-09-27T06:37:00Z", "06:37", "6:37 AM"},
		{"2026-09-27T21:37:00Z", "21:37", "9:37 PM"},
		{"2026-09-27T00:00:00Z", "00:00", "12:00 AM"},
		{"2026-09-27T12:00:00Z", "12:00", "12:00 PM"},
	}
	for _, each := range cases {
		at := instant(t, each.at)
		if got := Text(at, TwentyFourHour); got != each.twentyFour {
			t.Errorf("%s 24-hour: got %q, want %q", each.at, got, each.twentyFour)
		}
		if got := Text(at, TwelveHour); got != each.twelve {
			t.Errorf("%s 12-hour: got %q, want %q", each.at, got, each.twelve)
		}
	}
}

// An unknown format reads as 24-hour, the default; it is written as one too.
func TestAnUnknownFormatIsTwentyFourHour(t *testing.T) {
	t.Parallel()
	for _, given := range []Format{"", "13h", TwentyFourHour} {
		if got := Normalise(given); got != TwentyFourHour {
			t.Errorf("Normalise(%q) = %q; want 24h", given, got)
		}
	}
	if got := Text(instant(t, "2026-09-27T06:37:00Z"), "13h"); got != "06:37" {
		t.Errorf("an unknown format writes as 24-hour: got %q", got)
	}
	if !slices.Equal(Formats, []Format{TwentyFourHour, TwelveHour}) {
		t.Errorf("Formats = %v; want 24h then 12h", Formats)
	}
}

func TestZoneMarkForNumericForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		abbreviation string
		offset       int
		want         string
	}{
		{"EDT", -4 * secondsPerHour, "EDT"},
		{"-03", -3 * secondsPerHour, "UTC-3"},
		{"+0545", 5*secondsPerHour + 45*secondsPerMinute, "UTC+5:45"},
		{"-0930", -(9*secondsPerHour + 30*secondsPerMinute), "UTC-9:30"},
		{"+00", 0, "UTC"},
		{"", 0, "UTC"},
		{"+1345", 13*secondsPerHour + 45*secondsPerMinute, "UTC+13:45"},
	}
	for _, each := range cases {
		if got := ZoneMark(each.abbreviation, each.offset); got != each.want {
			t.Errorf("ZoneMark(%q, %d) = %q, want %q", each.abbreviation, each.offset, got, each.want)
		}
	}
}

func TestNextRefreshIsTheNextMinuteBoundary(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"2026-09-27T21:37:42.5Z": "2026-09-27T21:38:00Z",
		"2026-09-27T21:37:00Z":   "2026-09-27T21:38:00Z",
		"2026-12-31T23:59:59Z":   "2027-01-01T00:00:00Z",
	}
	for from, want := range cases {
		if got := NextRefresh(instant(t, from)); !got.Equal(instant(t, want)) {
			t.Errorf("after %s: got %s, want %s", from, got, want)
		}
	}
}

// London, Berlin, Tokyo, Melbourne then New York; UTC-10 after UTC+14; equal offsets keep their order.
func TestTheRibbonRunsEastFromGreenwich(t *testing.T) {
	t.Parallel()
	type place struct {
		name   string
		offset int
	}
	given := []place{
		{"New York", -4 * secondsPerHour}, {"Melbourne", 10 * secondsPerHour}, {"Tokyo", 9 * secondsPerHour},
		{"Berlin", 2 * secondsPerHour}, {"London", 1 * secondsPerHour}, {"Honolulu", -10 * secondsPerHour},
		{"Kiritimati", 14 * secondsPerHour}, {"Brighton", 1 * secondsPerHour}, {"Reykjavik", 0},
	}
	slices.SortStableFunc(given, func(a, b place) int { return EastFromGreenwich(a.offset, b.offset) })
	var names []string
	for _, each := range given {
		names = append(names, each.name)
	}
	want := []string{"Reykjavik", "London", "Brighton", "Berlin", "Tokyo", "Melbourne", "Kiritimati", "Honolulu", "New York"}
	if !slices.Equal(names, want) {
		t.Fatalf("order %v; want %v", names, want)
	}
}
