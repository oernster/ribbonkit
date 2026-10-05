package window

// The unpinned ribbon (FR-613 to FR-618): the facade's half of collapsing to the tab and opening
// from it. When is hover's to decide; this file carries the decision out on the window.

import (
	"fmt"
	"sync"
	"time"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/domain/hover"
	"github.com/oernster/ribbonkit/domain/placement"
)

// unpinned is the hover state with what the window shows because of it.
type unpinned struct {
	guard sync.Mutex
	state hover.State
	// shownOpen is whether the window shows the full ribbon rather than its tab.
	shownOpen bool
	// drawing is whether the page is drawing the full ribbon inside the tab, which the window grows
	// to once the page says it has drawn (RibbonDrawn) or drawWait has passed, whichever is first.
	drawing  bool
	stopDraw func() bool
	// full is the full ribbon's last arrangement, which the tab is cut from and opening returns to.
	// placed is false until the first arrangement, before which there is nothing to compare with.
	full   arranger.Arrangement
	placed bool
	// holds counts what keeps the ribbon as it is: an open panel, the ribbon's own menu.
	holds    int
	menuHeld bool
	// stop cancels the pending timer; nil when none is pending.
	stop func() bool
}

// drawWait is how long an opening ribbon waits for the page to say it has drawn before the window
// grows regardless, so a page that never answers cannot leave the ribbon as its tab. It is a ceiling
// for a page that has failed, not the expected wait: the page answers within a frame or two.
const drawWait = 250 * time.Millisecond

// pinned answers whether Pin ribbon is ticked: the choice, kept wherever the ribbon stands (FR-613).
func (a *Window) pinned() bool { return a.service.Choices().Pinned }

// pinnedAt answers whether a ribbon arranged as full behaves as pinned: chosen so; else standing
// flush against no edge along its orientation (FR-619). Pinned in effect, it shows in full.
func (a *Window) pinnedAt(full arranger.Arrangement) bool {
	return a.service.Choices().PinnedInEffect(full.Edge != "")
}

// collapsed answers whether the window is the ribbon's tab, which the page draws as a band (FR-614).
func (a *Window) collapsed() bool {
	a.unpin.guard.Lock()
	defer a.unpin.guard.Unlock()
	return !a.unpin.shownOpen && !a.unpin.drawing && !a.panelOpen.Load()
}

// endDrawing forgets an opening that has not yet grown the window. The caller holds the guard.
func (a *Window) endDrawing() {
	a.unpin.drawing = false
	if a.unpin.stopDraw != nil {
		a.unpin.stopDraw()
		a.unpin.stopDraw = nil
	}
}

// arrangeWindow places the window for the ribbon arranged as full: the full ribbon, else its tab
// while unpinned in effect and collapsed. Every placement of the ribbon comes through here, so the pin
// in effect is read afresh after each (FR-619). A ribbon that has just become unpinned in effect, as
// one dragged back onto an edge, starts from full: it collapses once the pointer has been off it for
// hover.Away rather than under the pointer.
func (a *Window) arrangeWindow(full arranger.Arrangement) error {
	a.unpin.guard.Lock()
	becameUnpinned := a.unpin.placed && a.pinnedAt(a.unpin.full) && !a.pinnedAt(full)
	// Standing onto or off an edge changes whether an unpinned ribbon is kept on top (FR-617).
	edgeChanged := a.unpin.placed && (a.unpin.full.Edge == "") != (full.Edge == "")
	if becameUnpinned {
		a.unpin.state = a.unpin.state.Expanded(a.now())
	}
	a.unpin.full, a.unpin.placed = full, true
	open := a.pinnedAt(full) || a.unpin.state.Open()
	a.unpin.shownOpen = open
	a.endDrawing()
	a.unpin.guard.Unlock()
	err := a.showArranged(full, open)
	if edgeChanged {
		a.applyAlwaysOnTop()
	}
	if becameUnpinned {
		a.changeHover(func(state hover.State, _ time.Time) hover.State { return state })
	}
	return err
}

// showArranged puts the window at full; else at the tab cut from it (FR-614). An unpinned ribbon
// keeps the tab's frame when full as well: giving Wails' frame back as it opened had Windows paint a
// caption and a close button over it for a frame or two (measured 2026-09-29), so only a ribbon
// pinned in effect and a panel wear Wails' frame.
func (a *Window) showArranged(full arranger.Arrangement, open bool) error {
	if open {
		a.report("framing the full ribbon", a.tabFrame(!a.pinnedAt(full)))
		at, size, _ := windowOf(full)
		_, ribbon, pullOut, shown := a.pullOutLayout()
		return a.placeShaped(at, size, placement.Shape(size, ribbon, pullOut, shown))
	}
	tab, err := a.service.Collapsed(full)
	if err != nil {
		return err
	}
	a.report("taking the frame off the tab", a.tabFrame(true))
	return a.placeWhole(tab.At, tab.Size)
}

// ribbonAt answers where the full ribbon stands: where it was last arranged while the window is its
// tab, so a change made while collapsed is fitted from the ribbon's place rather than the tab's.
func (a *Window) ribbonAt() (placement.Point, error) {
	a.unpin.guard.Lock()
	collapsed, at := !a.unpin.shownOpen, a.unpin.full.At
	a.unpin.guard.Unlock()
	if collapsed {
		return at, nil
	}
	window, err := a.position()
	return a.ribbonFromWindow(window), err
}

// pointerMoved hears the pointer come onto the ribbon or go off it (FR-615, FR-616).
func (a *Window) pointerMoved(arrived bool) {
	if a.pinned() {
		return
	}
	a.changeHover(func(state hover.State, now time.Time) hover.State {
		if arrived {
			return state.Arrived(now)
		}
		return state.Left(now)
	})
}

// hold keeps the ribbon as it is until the matching release (FR-616); holds nest.
func (a *Window) hold(expand bool) {
	a.changeHover(func(state hover.State, now time.Time) hover.State {
		a.unpin.holds++
		if expand {
			return state.Held().Expanded(now)
		}
		return state.Held()
	})
}

// release ends one hold; the last one lets the ribbon collapse again.
func (a *Window) release() {
	a.changeHover(func(state hover.State, now time.Time) hover.State {
		if a.unpin.holds == 0 {
			return state
		}
		a.unpin.holds--
		if a.unpin.holds > 0 {
			return state
		}
		return state.Released(now)
	})
}

// menuShown holds the ribbon while its own menu is open; menuClosed releases it (FR-616).
func (a *Window) menuShown() {
	a.unpin.guard.Lock()
	already := a.unpin.menuHeld
	a.unpin.menuHeld = true
	a.unpin.guard.Unlock()
	if !already {
		a.hold(false)
	}
}

func (a *Window) menuClosed() {
	a.unpin.guard.Lock()
	held := a.unpin.menuHeld
	a.unpin.menuHeld = false
	a.unpin.guard.Unlock()
	if held {
		a.release()
	}
}

// changeHover applies change to the hover state at the present moment, schedules the timer for what
// it now awaits, then opens or collapses the window where the state says so. While a panel stands the
// window is that panel, so the ribbon takes its form as the panel closes instead.
//
// Opening tells the page first and grows the window only once the page has drawn the full ribbon
// (grow). Growing first showed the tab's band stretched over the whole window until the page caught
// up (measured 2026-09-29). Collapsing shrinks the window first, which hides the change.
func (a *Window) changeHover(change func(hover.State, time.Time) hover.State) {
	a.unpin.guard.Lock()
	now := a.now()
	a.unpin.state = change(a.unpin.state, now)
	if a.unpin.stop != nil {
		a.unpin.stop()
		a.unpin.stop = nil
	}
	if due, pending := a.unpin.state.Due(); pending {
		a.unpin.stop = a.after(due.Sub(now), a.hoverDue)
	}
	open := a.unpin.state.Open() || a.pinnedAt(a.unpin.full)
	opening := open && !a.unpin.shownOpen && !a.unpin.drawing && !a.panelOpen.Load()
	collapsing := !open && (a.unpin.shownOpen || a.unpin.drawing) && !a.panelOpen.Load()
	if opening {
		a.unpin.drawing = true
		a.unpin.stopDraw = a.after(drawWait, a.drawDue)
	}
	if collapsing {
		a.endDrawing()
		a.unpin.shownOpen = false
	}
	full := a.unpin.full
	a.unpin.guard.Unlock()
	if collapsing {
		a.report("collapsing the ribbon", a.showArranged(full, false))
	}
	if opening || collapsing {
		a.emit(eventRefresh)
	}
}

// RibbonDrawn is the page saying it has drawn the full ribbon, so an opening ribbon's window grows
// (FR-615). Said at any other time, it changes nothing.
func (a *Window) RibbonDrawn() { a.grow() }

// grow gives the window the full ribbon the page has drawn, once per opening. A panel opened
// meanwhile is the window now; closing it places the ribbon.
func (a *Window) grow() {
	a.unpin.guard.Lock()
	drawing := a.unpin.drawing && !a.panelOpen.Load()
	a.endDrawing()
	if drawing {
		a.unpin.shownOpen = true
	}
	full := a.unpin.full
	a.unpin.guard.Unlock()
	if drawing {
		a.report("opening the ribbon", a.showArranged(full, true))
	}
}

// hoverDue makes the hover change that has fallen due; drawDue grows a ribbon whose page has not
// said it has drawn within drawWait. Each runs on its timer's own goroutine.
func (a *Window) hoverDue() {
	a.onTimer(func() {
		a.changeHover(func(state hover.State, now time.Time) hover.State { return state.At(now) })
	})
}

func (a *Window) drawDue() { a.onTimer(a.grow) }

// onTimer runs do on a timer's goroutine, where a panic is caught and logged rather than ending
// the application.
func (a *Window) onTimer(do func()) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(a.log, "recovered from %v while opening or collapsing the ribbon\n", failure)
		}
	}()
	do()
}

// setPinned pins or unpins the ribbon (FR-613). Pinned, it is shown in full with nothing watching the
// pointer; unpinned, it stays on top (FR-617) and collapses once the pointer has been away for
// hover.Away.
func (a *Window) setPinned(on bool) error {
	err := a.service.SetPinned(on)
	a.applyAlwaysOnTop()
	pinned := a.pinned()
	a.unpin.guard.Lock()
	if pinned {
		a.unpin.state = hover.State{}
	} else {
		a.unpin.state = hover.Unpinned(a.now(), false)
	}
	againstNone := a.unpin.full.Edge == ""
	a.unpin.guard.Unlock()
	if !pinned && againstNone {
		// Unticked away from every edge, the ribbon goes to the edge it last stood against (FR-613).
		a.placeBy("putting the ribbon against its last edge", a.service.ToLastEdge)
	}
	a.changeHover(func(state hover.State, _ time.Time) hover.State { return state })
	// A ribbon already shown in full stays so while trading frames: Wails' pinned, the tab's unpinned.
	a.unpin.guard.Lock()
	full, reframe := a.unpin.full, a.unpin.shownOpen && !a.panelOpen.Load()
	a.unpin.guard.Unlock()
	if reframe {
		a.report("framing the ribbon for its pin", a.showArranged(full, true))
	}
	a.trackPointer(!pinned && a.visible.Load())
	return err
}

// trackPointer starts or stops the desktop reporting the pointer, which is wanted only while an
// unpinned ribbon is shown.
func (a *Window) trackPointer(on bool) {
	if a.ribbon != 0 {
		a.watchPointer(on)
	}
}
