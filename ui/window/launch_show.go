package window

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oernster/ribbonkit/domain/placement"
)

// sizeWait is how long a launched ribbon whose page is ready waits for the page's reports that size
// it before it is shown anyway, so a page that never reports still shows a ribbon.
const sizeWait = time.Second

// launchShow holds what the first showing of a launched ribbon waits for: the page ready to show,
// its scale applied to the window and its widest text measured, the two reports that size the
// window after it loads. Shown at domReady alone, the window appeared at one size, then grew; at a
// fractional KDE scale WebKitGTK then often kept painting the size it was first shown at, leaving
// the ribbon cut off until the page next changed (measured 2026-10-02 on Plasma at 150 percent:
// shown at once, 8 of 16 launches were cut off to the first size; shown once scaled but before the
// widths, 13 of 16 were cut to the width of that moment).
type launchShow struct {
	ready    atomic.Bool
	scaled   atomic.Bool
	measured atomic.Bool
	shown    atomic.Bool
}

// pageReady records that the page can be shown, then shows the ribbon if it is sized; else it waits
// sizeWait for that.
func (a *Window) pageReady() {
	a.launch.ready.Store(true)
	if !a.showLaunched() {
		a.after(sizeWait, a.sizeDue)
	}
}

// pageScaled records that the page's scale has been applied to the window.
func (a *Window) pageScaled() {
	a.launch.scaled.Store(true)
	a.showLaunched()
}

// pageMeasured records that the page's widest text has been applied to the window.
func (a *Window) pageMeasured() {
	a.launch.measured.Store(true)
	a.showLaunched()
}

// sizeDue shows a ready ribbon whose page has not finished sizing it within sizeWait.
func (a *Window) sizeDue() {
	a.launch.scaled.Store(true)
	a.launch.measured.Store(true)
	a.showLaunched()
}

// showLaunched shows the launched ribbon once, when the page is ready and has sized it; it answers
// whether it has been shown.
func (a *Window) showLaunched() bool {
	if !a.launch.ready.Load() || !a.launch.scaled.Load() || !a.launch.measured.Load() {
		return false
	}
	if a.launch.shown.CompareAndSwap(false, true) {
		a.show()
		a.keepLaunchedPlace()
	}
	return true
}

// keepLaunchedPlace puts the window back where it was placed when the desktop showed it somewhere
// else (FR-403, FR-405). GNOME may place a newly shown window by its own rule, ignoring where it was
// put while hidden: measured 2026-10-04 on Ubuntu, placed at 1252,358 and shown at 248,72, whose
// settling was then stored as the user's drag. Placed again at once, before that move settles, the
// move is not taken for one.
func (a *Window) keepLaunchedPlace() {
	at, err := a.position()
	if err != nil {
		a.report("reading where the ribbon was shown", err)
		return
	}
	fmt.Fprintf(a.log, "launch: shown, the window stands at %v (NFR-O-1)\n", at)
	want, known := a.lastPlaced.read()
	if !known || at == want {
		return
	}
	fmt.Fprintf(a.log, "launch: the desktop showed the window at %v, not %v; placing it again\n", at, want)
	a.unpin.guard.Lock()
	full := a.unpin.full
	a.unpin.guard.Unlock()
	a.report("placing the ribbon again", a.arrangeWindow(full))
}

// placedWindow is where the window's top-left corner was last put.
type placedWindow struct {
	guard sync.Mutex
	at    placement.Point
	known bool
}

// note records at as where the window was last put.
func (p *placedWindow) note(at placement.Point) {
	p.guard.Lock()
	defer p.guard.Unlock()
	p.at, p.known = at, true
}

// read answers where the window was last put; false before it has been.
func (p *placedWindow) read() (placement.Point, bool) {
	p.guard.Lock()
	defer p.guard.Unlock()
	return p.at, p.known
}
