//go:build linux || darwin

package desktop

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
)

// errUnknownWindow is answered for a Window this package never handed out.
var errUnknownWindow = errors.New("that window was never found")

// Off Windows the desktop's windows are C pointers, a GtkWindow or an NSWindow, which a Window
// cannot carry as a number without the garbage collector's rules being bent, so each one found is
// kept here under the Window handed out for it. Where Place last put each window is kept beside it,
// so the end of a move can be told from the window arriving where it was placed.
var (
	known      sync.Mutex
	windows    = map[Window]unsafe.Pointer{}
	placed     = map[Window]placement.Point{}
	lastWindow Window
)

// placedAt answers where Place last put ribbon; false before it has.
func placedAt(ribbon Window) (placement.Point, bool) {
	known.Lock()
	defer known.Unlock()
	at, ok := placed[ribbon]
	return at, ok
}

// notePlaced records that Place is putting ribbon at at.
func notePlaced(ribbon Window, at placement.Point) {
	known.Lock()
	defer known.Unlock()
	placed[ribbon] = at
}

// remember answers the Window for window, handing out a new one the first time it is seen.
func remember(window unsafe.Pointer) Window {
	known.Lock()
	defer known.Unlock()
	for handed, held := range windows {
		if held == window {
			return handed
		}
	}
	lastWindow++
	windows[lastWindow] = window
	return lastWindow
}

// forget drops a Window whose native window is gone.
func forget(ribbon Window) {
	known.Lock()
	defer known.Unlock()
	delete(windows, ribbon)
	delete(placed, ribbon)
}

// findWith asks find for the ribbon's window until it answers one or findAttempts run out, then
// answers the Window for it, title naming it in the error when none was found. Wails creates the
// window as it starts, so it may not exist yet.
func findWith(title string, find func() unsafe.Pointer) (Window, error) {
	for range findAttempts {
		if found := find(); found != nil {
			return remember(found), nil
		}
		time.Sleep(findPause)
	}
	return 0, fmt.Errorf("%w: no window titled %s", ErrRibbonNotFound, title)
}

// pointerOf answers the native window a Window stands for.
func pointerOf(ribbon Window) (unsafe.Pointer, error) {
	known.Lock()
	defer known.Unlock()
	if window, ok := windows[ribbon]; ok {
		return window, nil
	}
	return nil, errUnknownWindow
}
