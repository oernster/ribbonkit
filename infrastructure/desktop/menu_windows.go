package desktop

import (
	"sync/atomic"
	"unsafe"

	"github.com/oernster/ribbonkit/application/menus"
)

// wmShowMenu asks the desktop's thread to show the menu waiting in pending.
const wmShowMenu = wmApp + 2

// pendingMenu holds the items a ShowMenu call handed over until the desktop's thread shows them.
type pendingMenu struct {
	items atomic.Pointer[[]menus.Item]
}

// ShowMenu shows items as a native menu at the cursor, from any goroutine: the ribbon's right-click
// menu (FR-108). A native menu is not clipped by the ribbon's small window, as one drawn in the page
// would be. The choice arrives as an EventMenu like the tray's.
func (d *Desktop) ShowMenu(items []menus.Item) {
	d.pending.items.Store(&items)
	if window := d.posted.Load(); window != 0 {
		_, _, _ = procPostMessage.Call(window, wmShowMenu, 0, 0)
	}
}

// showPending shows the menu ShowMenu handed over. It runs on the desktop's thread.
func (d *Desktop) showPending() {
	if items := d.pending.items.Swap(nil); items != nil {
		d.track(*items)
	}
}

// track shows items as a popup menu at the cursor and reports the choice (FR-502, FR-108). It runs
// on the desktop's thread.
func (d *Desktop) track(items []menus.Item) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	// Destroying a menu destroys every submenu attached to it.
	defer func() { _, _, _ = procDestroyMenu.Call(menu) }()
	next := 0
	fill(menu, items, &next)
	var cursor point
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	// A popup menu dismisses on a click elsewhere only when its owner is foreground first; it closes
	// cleanly only after a null message. Both are documented quirks.
	_, _, _ = procSetForegroundWindow.Call(uintptr(d.window))
	chosen, _, _ := procTrackPopupMenu.Call(menu, tpmRightButton|tpmNonotify|tpmReturnCmd,
		uintptr(cursor.x), uintptr(cursor.y), 0, uintptr(d.window), 0)
	_, _, _ = procPostMessage.Call(uintptr(d.window), wmNull, 0, 0)
	if action, ok := chosenAction(items, int(chosen)); ok {
		d.send(Event{Kind: EventMenu, Action: action})
	}
	d.send(Event{Kind: EventMenuClosed})
}

// fill appends items to menu, a submenu for each item holding children (FR-508). Identifiers are
// given out depth first from next, the order numbered answers them in.
func fill(menu uintptr, items []menus.Item, next *int) {
	for _, item := range items {
		if separatedBefore(item) {
			_, _, _ = procAppendMenu.Call(menu, mfSeparator, 0, 0)
		}
		if len(item.Children) > 0 {
			sub, _, _ := procCreatePopupMenu.Call()
			if sub == 0 {
				continue
			}
			fill(sub, item.Children, next)
			_, _, _ = procAppendMenu.Call(menu, mfPopup, sub, utf16Pointer(item.Label))
			continue
		}
		flags := uintptr(mfString)
		if item.Checkable && item.Checked {
			flags |= mfChecked
		}
		_, _, _ = procAppendMenu.Call(menu, flags, uintptr(menuIDBase+*next), utf16Pointer(item.Label))
		*next++
	}
}

// chosenAction answers the action of the menu identifier Windows answered; false for none.
// Identifiers start at menuIDBase, since Windows answers zero for no choice.
func chosenAction(items []menus.Item, id int) (menus.Action, bool) {
	return actionAt(items, id-menuIDBase)
}
