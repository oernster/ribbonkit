package desktop

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>

// seat_pointer answers the default seat's pointer, else NULL when GDK holds no display.
static GdkDevice *seat_pointer(void)
{
    GdkDisplay *display = gdk_display_get_default();
    if (display == NULL) {
        return NULL;
    }
    return gdk_seat_get_pointer(gdk_display_get_default_seat(display));
}

// pointer_at puts where the pointer is in x, y, in GTK's units on the root window (the units
// gtk_window_get_position answers the ribbon's corner in); FALSE when it cannot be read.
static gboolean pointer_at(int *x, int *y)
{
    GdkDevice *pointer = seat_pointer();
    if (pointer == NULL) {
        return FALSE;
    }
    gdk_device_get_position(pointer, NULL, x, y);
    return TRUE;
}

// pointer_warp moves the pointer to x, y in GTK's units on the root window; for the tests alone.
static void pointer_warp(int x, int y)
{
    GdkDevice *pointer = seat_pointer();
    if (pointer == NULL) {
        return;
    }
    gdk_device_warp(pointer, gdk_display_get_default_screen(gdk_device_get_display(pointer)), x, y);
    gdk_display_sync(gdk_device_get_display(pointer));
}
*/
import "C"

import (
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/gtkmain"
)

// Cursor answers where the pointer is in GTK's units, the units the ribbon is placed in; also
// whether it could be read. The corner grip's drag reads it rather than the page's pointer events,
// which were measured on Windows to jump backwards while the window was resized under them
// (FR-623). It is read on GTK's main loop.
func Cursor() (placement.Point, bool) {
	var x, y C.int
	var read C.gboolean
	gtkmain.Do(func() { read = C.pointer_at(&x, &y) })
	return placement.Point{X: int(x), Y: int(y)}, read != 0
}

// warpPointer moves the pointer to at, in the same units Cursor answers; for the tests alone.
func warpPointer(at placement.Point) {
	gtkmain.Do(func() { C.pointer_warp(C.int(at.X), C.int(at.Y)) })
}
