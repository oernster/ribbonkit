package desktop

/*
#cgo pkg-config: gtk+-3.0 x11
#include <gtk/gtk.h>
#include <gdk/gdkx.h>
#include <X11/Xatom.h>

// The opacities a veiled window is mapped at and given back; the property value the window manager
// reads as fully opaque.
#define VEILED_OPACITY 0.0
#define SHOWN_OPACITY 1.0
#define OPAQUE_CARDINAL 0xffffffffUL

// The mark a veiled window carries until it is next mapped.
static const char *veil_mark = "ribbonkit-veiled";

// veil_on_map handles a top-level window being mapped while GTK's main loop runs main_level deep.
// Wails maps its window with gtk_widget_show_all before the loop starts, even when told to start
// hidden (window.go, Run, in v2.12.0 to v2.16.0), so a window mapped at level 0 is veiled; any
// deliberate show comes through the loop, where a veiled window is given its opacity back.
//
// The opacity is set on the GDK window, which the window manager reads. gtk_widget_set_opacity
// would not do: for a top-level window with an RGBA visual, as Wails' translucent window has, GTK
// skips the window's opacity and fades the widget as it paints (gtk_widget_update_alpha, GTK
// 3.24), which cannot hide a window not yet painted (measured 2026-10-06: no property was set).
static void veil_on_map(GtkWidget *widget, guint main_level)
{
    if (!GTK_IS_WINDOW(widget) || gtk_window_get_window_type(GTK_WINDOW(widget)) != GTK_WINDOW_TOPLEVEL) {
        return;
    }
    GdkWindow *window = gtk_widget_get_window(widget);
    if (window == NULL) {
        return;
    }
    if (main_level == 0) {
        gdk_window_set_opacity(window, VEILED_OPACITY);
        g_object_set_data(G_OBJECT(widget), veil_mark, GINT_TO_POINTER(TRUE));
        return;
    }
    if (g_object_get_data(G_OBJECT(widget), veil_mark) != NULL) {
        gdk_window_set_opacity(window, SHOWN_OPACITY);
        g_object_set_data(G_OBJECT(widget), veil_mark, NULL);
    }
}

static gboolean veil_hook(GSignalInvocationHint *hint, guint count, const GValue *params, gpointer data)
{
    veil_on_map(GTK_WIDGET(g_value_get_object(&params[0])), gtk_main_level());
    return TRUE;
}

// veil_install hooks every widget's map signal once. The widget class is referenced first, since a
// signal cannot be looked up before its class exists; the reference is kept so the hook outlives
// the call.
static void veil_install(void)
{
    static gboolean installed = FALSE;
    if (installed) {
        return;
    }
    g_type_class_ref(GTK_TYPE_WIDGET);
    g_signal_add_emission_hook(g_signal_lookup("map", GTK_TYPE_WIDGET), 0, veil_hook, NULL, NULL);
    installed = TRUE;
}

// window_manager_opacity answers the opacity the window manager is asked to show widget at, read
// from its _NET_WM_WINDOW_OPACITY property, where no property means opaque; -1 off X11 or where the
// property cannot be read.
static double window_manager_opacity(GtkWidget *widget)
{
    GdkWindow *window = gtk_widget_get_window(widget);
    if (window == NULL || !GDK_IS_X11_DISPLAY(gdk_window_get_display(window))) {
        return -1;
    }
    GdkDisplay *display = gdk_window_get_display(window);
    Atom property = gdk_x11_get_xatom_by_name_for_display(display, "_NET_WM_WINDOW_OPACITY");
    Atom type;
    int format;
    unsigned long count, after;
    unsigned char *data = NULL;
    if (XGetWindowProperty(GDK_DISPLAY_XDISPLAY(display), GDK_WINDOW_XID(window), property, 0, 1, False,
                           XA_CARDINAL, &type, &format, &count, &after, &data) != Success) {
        return -1;
    }
    double got = SHOWN_OPACITY;
    if (data != NULL && count == 1) {
        got = (double)(*(unsigned long *)data) / (double)OPAQUE_CARDINAL;
    }
    if (data != NULL) {
        XFree(data);
    }
    return got;
}
*/
import "C"

import "unsafe"

// veilEarlyMaps keeps the window Wails shows by mistake from being seen. Wails maps its window at
// its default size before GTK's main loop starts, StartHidden or not; under Xwayland in the
// Flatpak it showed as a black window over most of the screen at login (measured 2026-10-06). A
// window mapped before the loop runs is mapped transparent; its next map, the ribbon's deliberate
// showing, makes it opaque again.
func veilEarlyMaps() { C.veil_install() }

// mapAtLevel applies the veil's rule to ribbon as though it were mapped with GTK's main loop
// running level deep, for the tests: the loop is always running by the time a test can map.
func mapAtLevel(ribbon Window, level uint) error {
	return onWindow(ribbon, func(window *C.GtkWindow) {
		C.veil_on_map((*C.GtkWidget)(unsafe.Pointer(window)), C.guint(level))
	})
}

// shownOpacity answers the opacity the window manager is asked to show ribbon at, 0 for transparent
// to 1 for opaque; false off X11, where there is no such property to read.
func shownOpacity(ribbon Window) (float64, bool, error) {
	var got float64
	err := onWindow(ribbon, func(window *C.GtkWindow) {
		got = float64(C.window_manager_opacity((*C.GtkWidget)(unsafe.Pointer(window))))
	})
	return got, got >= 0, err
}
