package placement

import (
	"errors"
	"math"
)

// ErrNegativeLength is answered when a length that cannot be negative is given as one.
var ErrNegativeLength = errors.New("a length cannot be negative")

// The page the ribbon shows is laid out in DIP and drawn at some number of window pixels to each
// DIP. That number is usually the display's DPI over BaseDPI; not always, though. Windows' text size
// enlarges the whole page while the display's DPI stays as it was (measured 2026-09-28: at Text
// size 125 percent on a 96 DPI display the page was drawn at 1.25). So a window is sized by the
// scale the page is really drawn at, which is what these convert with.

// floatSlack absorbs the error of a float product, so a length that is whole in exact arithmetic
// is not rounded one pixel past it.
const floatSlack = 1e-9

// PerDIPOf answers how many pixels a display at dpi gives each DIP.
func PerDIPOf(dpi int) float64 {
	if dpi <= 0 {
		dpi = BaseDPI
	}
	return float64(dpi) / BaseDPI
}

// PixelsOf answers length in DIP as whole pixels at perDIP pixels to the DIP, rounded up so a page
// drawn at that scale is never cut off.
func PixelsOf(length int, perDIP float64) int {
	return int(math.Ceil(float64(length)*perDIP - floatSlack))
}

// DIPOf answers pixels as whole DIP at perDIP pixels to the DIP, rounded down so what is sized to
// fit in it never overflows the pixels.
func DIPOf(pixels int, perDIP float64) int {
	return int(math.Floor(float64(pixels)/perDIP + floatSlack))
}
