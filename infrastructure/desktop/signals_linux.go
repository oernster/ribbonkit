package desktop

// GTK's side of the Linux desktop: connecting the signals it calls back on and building the
// ribbon's popup menu. The C half is in desktop_linux.c; the callbacks it reaches are in
// callbacks_unix.go.

/*
#cgo pkg-config: gtk+-3.0
#include <stdlib.h>
#include <gtk/gtk.h>
void desktop_watch(GtkWindow *ribbon, guintptr handle);
GtkWidget *menu_new(void);
void menu_add_item(GtkWidget *menu, const char *label, gboolean checkable, gboolean checked, gboolean enabled, int index);
GtkWidget *menu_add_submenu(GtkWidget *menu, const char *label);
void menu_add_separator(GtkWidget *menu);
void desktop_popup(GtkWindow *ribbon, GtkWidget *menu);
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"

	"github.com/oernster/ribbonkit/application/menus"
)

// watchWindow connects the ribbon's moves and the screen's changes of display to the desktop that
// handle holds.
func watchWindow(ribbon Window, handle cgo.Handle) error {
	return onWindow(ribbon, func(window *C.GtkWindow) { C.desktop_watch(window, C.guintptr(handle)) })
}

// popUp shows items over ribbon at the pointer, their choice going to d.
func popUp(ribbon Window, d *Desktop, items []menus.Item) error {
	showing(d, items)
	return onWindow(ribbon, func(window *C.GtkWindow) {
		menu := C.menu_new()
		next := 0
		build(menu, items, &next)
		C.desktop_popup(window, menu)
	})
}

// build adds items to menu, a submenu for each item holding children. Numbers are given out depth
// first from next, the order actionAt reads them in.
func build(menu *C.GtkWidget, items []menus.Item, next *int) {
	for _, item := range items {
		if separatedBefore(item) {
			C.menu_add_separator(menu)
		}
		label := C.CString(item.Label)
		if len(item.Children) > 0 {
			sub := C.menu_add_submenu(menu, label)
			C.free(unsafe.Pointer(label))
			build(sub, item.Children, next)
			continue
		}
		C.menu_add_item(menu, label, gboolean(item.Checkable), gboolean(item.Checked), gboolean(!item.Disabled), C.int(*next))
		C.free(unsafe.Pointer(label))
		*next++
	}
}

// gboolean answers GTK's truth value for b.
func gboolean(b bool) C.gboolean {
	if b {
		return C.TRUE
	}
	return C.FALSE
}
