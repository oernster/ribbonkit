// The C half of signals_linux.go: GTK signal handlers and the ribbon's popup menu. Everything here
// runs on GTK's main loop.

#include <gtk/gtk.h>
#include <gdk/gdkx.h>
#include "_cgo_export.h"

// The key each menu item carries its number under.
#define MENU_INDEX_KEY "ribbonkit-index"

// The secondary button, whose press opens the ribbon's menu.
#define SECONDARY_BUTTON 3

// on_configure reports where the ribbon stands after the window manager moved or sized it. The
// event's own coordinates are relative to a parent, which may be the window manager's frame, so the
// position is read from the window instead.
static gboolean on_configure(GtkWidget *widget, GdkEventConfigure *event, gpointer data)
{
    gint x = 0, y = 0;
    gtk_window_get_position(GTK_WINDOW(widget), &x, &y);
    desktopMoved((GoUintptr)data, x, y);
    return FALSE;
}

static void on_monitors_changed(GdkScreen *screen, gpointer data)
{
    desktopDisplaysChanged((GoUintptr)data);
}

// on_crossing reports the pointer coming onto the ribbon's window or leaving it (FR-615, FR-616). A
// crossing into or out of a child, the web view, leaves it on the window. A crossing made by a grab,
// as a drag or a popup menu takes the pointer, is not a departure: the drag holds the ribbon and the
// menu reports its own closing.
static gboolean on_crossing(GtkWidget *widget, GdkEventCrossing *event, gpointer data)
{
    if (event->mode == GDK_CROSSING_GRAB) {
        return FALSE;
    }
    desktopPointer((GoUintptr)data, event->type == GDK_ENTER_NOTIFY || event->detail == GDK_NOTIFY_INFERIOR);
    return FALSE;
}

void desktop_watch(GtkWindow *ribbon, guintptr handle)
{
    g_signal_connect(ribbon, "configure-event", G_CALLBACK(on_configure), (gpointer)handle);
    gtk_widget_add_events(GTK_WIDGET(ribbon), GDK_ENTER_NOTIFY_MASK | GDK_LEAVE_NOTIFY_MASK);
    g_signal_connect(ribbon, "enter-notify-event", G_CALLBACK(on_crossing), (gpointer)handle);
    g_signal_connect(ribbon, "leave-notify-event", G_CALLBACK(on_crossing), (gpointer)handle);
    g_signal_connect(gdk_screen_get_default(), "monitors-changed", G_CALLBACK(on_monitors_changed),
                     (gpointer)handle);
}

static void on_item_activate(GtkMenuItem *item, gpointer data)
{
    desktopMenuChosen(GPOINTER_TO_INT(g_object_get_data(G_OBJECT(item), MENU_INDEX_KEY)));
}

GtkWidget *menu_new(void)
{
    return gtk_menu_new();
}

// menu_add_item adds an item numbered index. A check item's state is set before its handler is
// connected, since setting it emits activate, which would report a choice nobody made.
void menu_add_item(GtkWidget *menu, const char *label, gboolean checkable, gboolean checked, int index)
{
    GtkWidget *item;
    if (checkable) {
        item = gtk_check_menu_item_new_with_label(label);
        gtk_check_menu_item_set_active(GTK_CHECK_MENU_ITEM(item), checked);
    } else {
        item = gtk_menu_item_new_with_label(label);
    }
    g_object_set_data(G_OBJECT(item), MENU_INDEX_KEY, GINT_TO_POINTER(index));
    g_signal_connect(item, "activate", G_CALLBACK(on_item_activate), NULL);
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), item);
}

GtkWidget *menu_add_submenu(GtkWidget *menu, const char *label)
{
    GtkWidget *item = gtk_menu_item_new_with_label(label);
    GtkWidget *sub = gtk_menu_new();
    gtk_menu_item_set_submenu(GTK_MENU_ITEM(item), sub);
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), item);
    return sub;
}

void menu_add_separator(GtkWidget *menu)
{
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), gtk_separator_menu_item_new());
}

// destroy_later destroys a closed menu once the loop is idle, after its chosen item has activated,
// then reports it closed, so the choice is heard first, as on Windows (FR-616).
static gboolean destroy_later(gpointer menu)
{
    gtk_widget_destroy(GTK_WIDGET(menu));
    desktopMenuClosed();
    return G_SOURCE_REMOVE;
}

static void on_menu_deactivate(GtkMenuShell *menu, gpointer data)
{
    g_idle_add(destroy_later, menu);
}

// secondary_press makes the press that opened the menu, as GTK would have seen it. The right-click
// reached the page, not GTK, so GTK has no event of its own. GTK keeps a menu open through the
// release that ends the click which opened it, judging that by the press's time; with no press, or
// one stamped at time zero, every release looks long after it and closes the menu the moment it
// opens. Measured 2026-09-28: with no press, "no trigger event for menu popup" on every attempt;
// with one at time zero, the menu hid on each release. The X server's own time settles it.
static GdkEvent *secondary_press(GdkWindow *window, GdkDevice *pointer, gint x, gint y)
{
    GdkEvent *press = gdk_event_new(GDK_BUTTON_PRESS);
    press->button.window = g_object_ref(window);
    press->button.send_event = TRUE;
    press->button.time = GDK_IS_X11_WINDOW(window) ? gdk_x11_get_server_time(window) : GDK_CURRENT_TIME;
    press->button.x = x;
    press->button.y = y;
    press->button.button = SECONDARY_BUTTON;
    gdk_event_set_device(press, pointer);
    return press;
}

// desktop_popup shows menu at the pointer over the ribbon.
void desktop_popup(GtkWindow *ribbon, GtkWidget *menu)
{
    GdkWindow *window = gtk_widget_get_window(GTK_WIDGET(ribbon));
    if (window == NULL) {
        gtk_widget_destroy(menu);
        return;
    }
    GdkDevice *pointer = gdk_seat_get_pointer(gdk_display_get_default_seat(gdk_window_get_display(window)));
    gint x = 0, y = 0;
    gdk_window_get_device_position(window, pointer, &x, &y, NULL);
    GdkRectangle at = {x, y, 1, 1};
    GdkEvent *press = secondary_press(window, pointer, x, y);
    gtk_menu_attach_to_widget(GTK_MENU(menu), GTK_WIDGET(ribbon), NULL);
    g_signal_connect(menu, "deactivate", G_CALLBACK(on_menu_deactivate), NULL);
    gtk_widget_show_all(menu);
    gtk_menu_popup_at_rect(GTK_MENU(menu), window, &at, GDK_GRAVITY_NORTH_WEST, GDK_GRAVITY_NORTH_WEST, press);
    gdk_event_free(press);
}
