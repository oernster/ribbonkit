package desktop

/*
#cgo pkg-config: gtk+-3.0
#include <stdlib.h>
#include <gtk/gtk.h>

// The test window's size before a test places it; placing it replaces both.
#define TEST_WINDOW_WIDTH 200
#define TEST_WINDOW_HEIGHT 100

// ribbon_toplevel answers the process's top-level window titled title, else NULL. Menus and
// tooltips are popups, never top-level, so they cannot be mistaken for it.
static GtkWindow *ribbon_toplevel(const char *title)
{
    GList *all = gtk_window_list_toplevels();
    GtkWindow *found = NULL;
    for (GList *each = all; each != NULL; each = each->next) {
        GtkWindow *window = GTK_WINDOW(each->data);
        if (gtk_window_get_window_type(window) == GTK_WINDOW_TOPLEVEL &&
            g_strcmp0(gtk_window_get_title(window), title) == 0) {
            found = window;
            break;
        }
    }
    g_list_free(all);
    return found;
}

static void ribbon_hide_from_taskbar(GtkWindow *window)
{
    gtk_window_set_skip_taskbar_hint(window, TRUE);
    gtk_window_set_skip_pager_hint(window, TRUE);
}

static gboolean ribbon_skips_taskbar(GtkWindow *window)
{
    return gtk_window_get_skip_taskbar_hint(window) && gtk_window_get_skip_pager_hint(window);
}

// ribbon_resize sizes the window. Wails makes its window non-resizable; GTK sizes such a window to
// what it requests, so the request is set as well as the size.
static void ribbon_resize(GtkWindow *window, int width, int height)
{
    gtk_widget_set_size_request(GTK_WIDGET(window), width, height);
    gtk_window_resize(window, width, height);
}

static int drag_threshold(void)
{
    gint threshold = 0;
    g_object_get(gtk_settings_get_default(), "gtk-dnd-drag-threshold", &threshold, NULL);
    return threshold;
}

// test_window makes a window set up as Wails sets up the ribbon's: titled, undecorated and not
// resizable, then shows it.
static GtkWindow *test_window(const char *title)
{
    GtkWindow *window = GTK_WINDOW(gtk_window_new(GTK_WINDOW_TOPLEVEL));
    gtk_window_set_title(window, title);
    gtk_window_set_decorated(window, FALSE);
    gtk_window_set_resizable(window, FALSE);
    gtk_window_set_default_size(window, TEST_WINDOW_WIDTH, TEST_WINDOW_HEIGHT);
    gtk_widget_show_all(GTK_WIDGET(window));
    return window;
}

static void test_window_destroy(GtkWindow *window) { gtk_widget_destroy(GTK_WIDGET(window)); }
*/
import "C"

import (
	"io"
	"time"
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/gtkmain"
)

// sizeLimit and sizePause bound the wait for a new size to take before the ribbon is moved.
const (
	sizeLimit = 500 * time.Millisecond
	sizePause = 10 * time.Millisecond
)

// gtkWindow answers the GTK window a Window stands for.
func gtkWindow(ribbon Window) (*C.GtkWindow, error) {
	window, err := pointerOf(ribbon)
	return (*C.GtkWindow)(window), err
}

// FindRibbon answers the ribbon's window. On Linux there is no window class to find it by, so
// class is not used: the ribbon is the process's top-level window titled name, the product's name,
// as Wails creates it.
func FindRibbon(_, name string) (Window, error) {
	title := C.CString(name)
	defer C.free(unsafe.Pointer(title))
	return findWith(name, func() unsafe.Pointer {
		var found *C.GtkWindow
		gtkmain.Do(func() { found = C.ribbon_toplevel(title) })
		return unsafe.Pointer(found)
	})
}

// HideFromTaskbar asks the window manager to leave the ribbon off the taskbar and the workspace
// switcher (FR-101).
func HideFromTaskbar(ribbon Window) error {
	return onWindow(ribbon, func(window *C.GtkWindow) { C.ribbon_hide_from_taskbar(window) })
}

// KeepOnDisplays does nothing on Linux, where the window manager carries a drag through and offers
// no say in it while it lasts. A ribbon dragged partly off every display is put back when the move
// ends, through the same path as on Windows (FR-404, FR-406).
func KeepOnDisplays(Window, io.Writer) error { return nil }

// Place moves and sizes the ribbon in GTK's units (FR-405).
func Place(ribbon Window, at placement.Point, size placement.Size) error {
	window, err := gtkWindow(ribbon)
	if err != nil {
		return err
	}
	notePlaced(ribbon, at)
	gtkmain.Do(func() { C.ribbon_resize(window, C.int(size.Width), C.int(size.Height)) })
	awaitSize(window, size)
	gtkmain.Do(func() { C.gtk_window_move(window, C.gint(at.X), C.gint(at.Y)) })
	return nil
}

// awaitSize waits, within sizeLimit, for window to take size. The window manager keeps a window on
// screen by the size it has when the move arrives, so a move sent before a shrink has landed is
// clamped as if the window were still large (measured 2026-09-28: a ribbon returning from a
// 560x760 panel was pushed to that panel's corner of the work area).
func awaitSize(window *C.GtkWindow, want placement.Size) {
	for deadline := time.Now().Add(sizeLimit); time.Now().Before(deadline); time.Sleep(sizePause) {
		var width, height C.gint
		gtkmain.Do(func() { C.gtk_window_get_size(window, &width, &height) })
		if int(width) == want.Width && int(height) == want.Height {
			return
		}
	}
}

// Position answers the ribbon's top-left corner in GTK's units.
func Position(ribbon Window) (placement.Point, error) {
	var x, y C.gint
	err := onWindow(ribbon, func(window *C.GtkWindow) { C.gtk_window_get_position(window, &x, &y) })
	return placement.Point{X: int(x), Y: int(y)}, err
}

// DragThreshold answers how far the pointer must move, in DIP, before a press becomes a drag: the
// desktop's own drag threshold, which GTK gives as one distance for both directions (FR-401).
func DragThreshold() placement.Size {
	var threshold C.int
	gtkmain.Do(func() { threshold = C.drag_threshold() })
	return placement.Size{Width: int(threshold), Height: int(threshold)}
}

// onWindow runs act on the GTK window ribbon stands for, on GTK's main loop.
func onWindow(ribbon Window, act func(*C.GtkWindow)) error {
	window, err := gtkWindow(ribbon)
	if err != nil {
		return err
	}
	gtkmain.Do(func() { act(window) })
	return nil
}

// skipsTaskbar answers whether the ribbon is marked to stay off the taskbar and switcher.
func skipsTaskbar(ribbon Window) (bool, error) {
	var skips bool
	err := onWindow(ribbon, func(window *C.GtkWindow) { skips = C.ribbon_skips_taskbar(window) != 0 })
	return skips, err
}

// size answers the ribbon's size in GTK's units.
func size(ribbon Window) (placement.Size, error) {
	var width, height C.gint
	err := onWindow(ribbon, func(window *C.GtkWindow) { C.gtk_window_get_size(window, &width, &height) })
	return placement.Size{Width: int(width), Height: int(height)}, err
}

// newTestWindow shows a window set up as Wails sets up the ribbon's, for the tests: nothing in the
// application makes one.
func newTestWindow(name string) Window {
	title := C.CString(name)
	defer C.free(unsafe.Pointer(title))
	var window *C.GtkWindow
	gtkmain.Do(func() { window = C.test_window(title) })
	return remember(unsafe.Pointer(window))
}

// closeTestWindow destroys a window newTestWindow made.
func closeTestWindow(ribbon Window) {
	_ = onWindow(ribbon, func(window *C.GtkWindow) { C.test_window_destroy(window) })
	forget(ribbon)
}
