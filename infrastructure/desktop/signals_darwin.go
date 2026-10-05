package desktop

// AppKit's side of the macOS desktop: watching the ribbon's moves and the displays, then building
// the menus. The Objective-C half is in desktop_darwin.m; the callbacks it reaches are in
// callbacks_unix.go.

/*
#include <stdlib.h>
#include <stdint.h>
void desktop_watch(void *ribbon, uintptr_t handle);
void *menu_new(void);
void menu_add_item(void *menu, const char *label, int checkable, int checked, int index);
void *menu_add_submenu(void *menu, const char *label);
void menu_add_separator(void *menu);
void desktop_popup(void *menu);
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"

	"github.com/oernster/ribbonkit/application/menus"
)

// watchWindow reports the ribbon's moves and the changes of display to the desktop handle holds.
func watchWindow(ribbon Window, handle cgo.Handle) error {
	return onWindow(ribbon, func(window unsafe.Pointer) { C.desktop_watch(window, C.uintptr_t(handle)) })
}

// popUp shows items at the pointer over ribbon, their choice going to d.
func popUp(ribbon Window, d *Desktop, items []menus.Item) error {
	showing(d, items)
	return onWindow(ribbon, func(unsafe.Pointer) {
		menu := C.menu_new()
		next := 0
		build(menu, items, &next)
		C.desktop_popup(menu)
	})
}

// build adds items to menu, a submenu for each item holding children. Numbers are given out depth
// first from next, the order actionAt reads them in. It runs on AppKit's main thread.
func build(menu unsafe.Pointer, items []menus.Item, next *int) {
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
		C.menu_add_item(menu, label, truth(item.Checkable), truth(item.Checked), C.int(*next))
		C.free(unsafe.Pointer(label))
		*next++
	}
}

// truth answers C's truth value for b.
func truth(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
