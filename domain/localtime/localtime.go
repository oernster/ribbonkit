// Package localtime holds what every ribbon showing places knows about their local time: how a time
// is written, the mark naming a zone, when the next minute begins and the order places run in, east
// from Greenwich.
//
// It takes every instant as an argument and every zone already resolved, so it reads no clock and
// holds no tz database of its own.
package localtime

import (
	"cmp"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"
)

// Format is how a time is written: 24-hour or 12-hour.
type Format string

// The two formats. The string values are what a settings file holds.
const (
	TwentyFourHour Format = "24h"
	TwelveHour     Format = "12h"
)

// Formats lists the formats in the order they are offered.
var Formats = []Format{TwentyFourHour, TwelveHour}

// Go reference layouts for the time in each format: two-digit hours in 24-hour time; unpadded hours
// with AM or PM in 12-hour time.
const (
	layoutTwentyFour = "15:04"
	layoutTwelve     = "3:04 PM"
)

// Clock arithmetic for the zone mark.
const (
	secondsPerHour   = 3600
	secondsPerMinute = 60
)

// offsetPrefix begins a zone mark written as an offset rather than an abbreviation.
const offsetPrefix = "UTC"

// Normalise answers format where it is one of Formats; TwentyFourHour otherwise, so a file holding an
// unknown word still draws.
func Normalise(format Format) Format {
	if format == TwelveHour {
		return TwelveHour
	}
	return TwentyFourHour
}

// Text answers the time local shows on its own clock, written in format: "06:37" or "6:37 AM",
// "12:00 AM" at midnight and "12:00 PM" at noon in 12-hour time.
func Text(local time.Time, format Format) string {
	if Normalise(format) == TwelveHour {
		return local.Format(layoutTwelve)
	}
	return local.Format(layoutTwentyFour)
}

// ZoneMark answers the text naming a zone beside a label: the abbreviation the tz database gives
// where it begins with a letter; otherwise "UTC" with the signed offset in hours, minutes added only
// when there are some. An offset of zero is "UTC" alone.
func ZoneMark(abbreviation string, offsetSeconds int) string {
	first, _ := utf8.DecodeRuneInString(abbreviation)
	if unicode.IsLetter(first) {
		return abbreviation
	}
	if offsetSeconds == 0 {
		return offsetPrefix
	}
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	hours := offsetSeconds / secondsPerHour
	minutes := offsetSeconds % secondsPerHour / secondsPerMinute
	if minutes == 0 {
		return fmt.Sprintf("%s%s%d", offsetPrefix, sign, hours)
	}
	return fmt.Sprintf("%s%s%d:%02d", offsetPrefix, sign, hours, minutes)
}

// NextRefresh answers the next minute boundary after instant, worked out from the instant itself
// rather than from the last refresh, so refreshes cannot drift.
func NextRefresh(instant time.Time) time.Time {
	return instant.Truncate(time.Minute).Add(time.Minute)
}

// EastFromGreenwich orders two places by their offsets from UTC in seconds, starting at Greenwich and
// going east round the world: every place level with or ahead of UTC before every place behind it,
// each group by ascending offset. So UTC-10 never comes before UTC+14, although the two keep the same
// time of day. It answers as cmp.Compare does, for slices.SortStableFunc, so places keeping the same
// offset keep their order.
func EastFromGreenwich(a, b int) int {
	if behind := a < 0; behind != (b < 0) {
		if behind {
			return 1
		}
		return -1
	}
	return cmp.Compare(a, b)
}
