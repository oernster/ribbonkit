package desktop

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>

// toolkit_scale answers GDK's window scale, the whole number GTK multiplies its units by. On X11 it
// is one number for the whole display, so the first monitor's is every monitor's; 1 when GDK holds
// no display to ask.
static int toolkit_scale(void)
{
    GdkDisplay *display = gdk_display_get_default();
    if (display == NULL) {
        return 1;
    }
    GdkMonitor *monitor = gdk_display_get_monitor(display, 0);
    if (monitor == NULL) {
        return 1;
    }
    return gdk_monitor_get_scale_factor(monitor);
}
*/
import "C"

import "github.com/oernster/ribbonkit/infrastructure/gtkmain"

// wholeScale is the toolkit scale of a window GTK does not scale itself.
const wholeScale = 1

// PixelsPerDIP answers how many window units the page's CSS pixel takes, given the page's
// devicePixelRatio and GTK's own window scale. GTK sizes the window in its units, device pixels over
// its window scale, a whole number. WebKitGTK draws the page at devicePixelRatio, which carries that
// whole scale and also the font DPI the desktop sets: KDE hands an X11 program a fractional display
// scale as font DPI alone (measured 2026-10-02 on Plasma at 150 percent: no window scaling factor
// in its XSETTINGS and no GDK_SCALE, font DPI 144, devicePixelRatio 1.5). So the window takes the ratio over GTK's scale; taking one unit for every
// CSS pixel left the page cut off. A scale below one is treated as GTK not scaling.
func PixelsPerDIP(pageRatio float64, toolkitScale int) float64 {
	if toolkitScale < wholeScale {
		toolkitScale = wholeScale
	}
	return pageRatio / float64(toolkitScale)
}

// ToolkitScale answers GDK's window scale, read on GTK's main loop.
func ToolkitScale() int {
	scale := wholeScale
	gtkmain.Do(func() { scale = int(C.toolkit_scale()) })
	return scale
}
