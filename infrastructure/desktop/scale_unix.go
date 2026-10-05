//go:build darwin

package desktop

// unitsPerCSSPixel is how many of AppKit's units the page's CSS pixel takes: AppKit sizes the
// window in points and scales the page with it, so one.
const unitsPerCSSPixel = 1

// PixelsPerDIP answers how many window units the page's CSS pixel takes, given the page's
// devicePixelRatio and the toolkit's own window scale. On macOS the window is sized in points, so
// the display's backing scale in devicePixelRatio is AppKit's business and every CSS pixel is one
// unit.
func PixelsPerDIP(float64, int) float64 { return unitsPerCSSPixel }

// ToolkitScale answers the toolkit's own window scale. AppKit's points already carry the backing
// scale, which PixelsPerDIP ignores on macOS, so one.
func ToolkitScale() int { return unitsPerCSSPixel }
