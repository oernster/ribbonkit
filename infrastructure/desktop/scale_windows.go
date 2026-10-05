package desktop

// PixelsPerDIP answers how many window pixels the page's CSS pixel takes, given the page's
// devicePixelRatio. On Windows the window is sized in physical pixels and WebView2 draws the page at
// its rasterization scale, the display's scale combined with the user's text size, which is what
// devicePixelRatio reports; so the two are the same. There is no toolkit scale to take out.
func PixelsPerDIP(pageRatio float64, _ int) float64 { return pageRatio }

// noToolkitScale is the toolkit scale on Windows, where the window is sized in physical pixels.
const noToolkitScale = 1

// ToolkitScale answers the toolkit's own window scale, which Windows does not have.
func ToolkitScale() int { return noToolkitScale }
