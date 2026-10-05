package window

// The window's own life: startup, showing, hiding, closing and what the desktop reports.

import (
	"context"
	"fmt"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/shell"
	"github.com/oernster/ribbonkit/domain/placement"
)

// startup takes the ribbon off the taskbar and puts it in place while it is still hidden, then
// starts listening to the desktop. Nothing here ends the run: a failure is logged and the ribbon
// opens wherever Wails put it.
func (a *Window) startup(ctx context.Context) {
	a.ctx.Store(&ctx)
	go a.watchForUpdates(ctx)
	ribbon, err := a.desktop.FindRibbon(a.product.WindowClass, a.product.App.Name)
	if err != nil {
		a.report("finding the ribbon", err)
		return
	}
	a.ribbon = ribbon
	a.report("hiding the taskbar button", a.desktop.HideFromTaskbar(ribbon))
	a.report("keeping the ribbon on its displays", a.desktop.KeepOnDisplays(ribbon, a.log))
	a.desktop.Watch(ribbon)
	a.report("placing the ribbon", a.placeLaunched())
	a.applyAlwaysOnTop()
	go a.listen()
}

// domReady shows the ribbon once the page has drawn and its scale is known, so it never appears
// blank and never grows while first painting (launch_show.go).
func (a *Window) domReady(context.Context) { a.pageReady() }

// beforeClose answers a request to close the ribbon, such as Alt+F4: it hides the ribbon and the
// application keeps running (FR-507). An Exit already decided passes through, as does any close
// while there is no tray icon to bring the ribbon back from.
func (a *Window) beforeClose(context.Context) bool {
	if a.quitting.Load() || !a.trayUp.Load() {
		return false
	}
	if a.service.CloseRequested() == menus.Hide {
		a.hide()
	}
	return true
}

func (a *Window) shutdown(context.Context) { a.desktop.Stop() }

// secondInstance answers a second launch by toggling the ribbon as the tray icon's click does, so
// one launcher button, such as a Stream Deck's, both shows and hides it (FR-506, Amendment 17).
func (a *Window) secondInstance() { a.toggle() }

// listen acts on what the desktop reports until it stops. A panic in one event is logged and the
// next is still heard, so one fault cannot leave a ribbon that reacts to nothing.
func (a *Window) listen() {
	for event := range a.desktop.Events() {
		a.handleSafely(event)
	}
}

func (a *Window) handleSafely(event shell.Event) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(a.log, "recovered from %v while handling desktop event %d\n", failure, event.Kind)
		}
	}()
	switch event.Kind {
	case shell.EventMenu:
		a.act(event.Action)
	case shell.EventIconClicked:
		a.toggle()
	case shell.EventMoveEnded:
		a.moved()
	case shell.EventDisplayChanged:
		fmt.Fprintln(a.log, "the displays changed")
		a.rearrange()
		a.emit(eventRefresh)
	case shell.EventTimeChanged, shell.EventResumed:
		fmt.Fprintf(a.log, "desktop event %d: refreshing\n", event.Kind)
		a.emit(eventRefresh)
	case shell.EventPointerArrived, shell.EventPointerLeft:
		a.pointerMoved(event.Kind == shell.EventPointerArrived)
	case shell.EventMenuClosed:
		a.menuClosed()
	}
}

// act carries out a menu action from the tray or the ribbon's own menu.
func (a *Window) act(action menus.Action) {
	switch action {
	case menus.Show:
		a.show()
	case menus.Hide:
		a.hide()
	case menus.Settings:
		a.show()
		a.emit(eventOpenPanel, openAtSettings)
	case menus.About:
		a.show()
		a.emit(eventOpenPanel, openAtAbout)
	case menus.Licence:
		a.show()
		a.emit(eventOpenPanel, openAtLicence)
	case menus.Updates:
		go a.checkForUpdate(a.wailsContext(), true)
	case menus.AlwaysOnTop:
		a.report("changing Always on top", a.SetAlwaysOnTop(!a.service.Choices().AlwaysOnTop))
		a.emit(eventRefresh)
	case menus.Pin:
		a.report("pinning the ribbon", a.setPinned(!a.pinned()))
		a.emit(eventRefresh)
	case menus.Exit:
		a.quitting.Store(true)
		if a.wailsContext() != nil {
			a.quit()
		}
	default:
		a.actOnChoice(action)
	}
}

// actOnChoice carries out a Position item (FR-408); any other action is one of the application's
// own, which it carries out itself (Config.Act).
func (a *Window) actOnChoice(action menus.Action) {
	if edge, ok := menus.EdgeOf(action); ok {
		a.toEdge(edge)
		return
	}
	a.actOn(action)
}

// toEdge puts the ribbon against edge of its display and shows it there (FR-408). While a panel is
// open the window is that panel, so the place is kept and the ribbon goes there as the panel closes.
// Before startup has found the ribbon there is nothing to move.
func (a *Window) toEdge(edge placement.Edge) {
	placed := a.placeBy("putting the ribbon against an edge", func(at placement.Point) (arranger.Arrangement, error) {
		return a.service.ToEdge(at, edge)
	})
	if placed {
		a.show()
	}
}

// placeBy places the ribbon where arrange answers for it as it stands, answering whether the window
// was placed. While a panel is open the window is that panel, so the place is kept and the ribbon goes
// there as the panel closes. Before startup has found the ribbon there is nothing to move.
func (a *Window) placeBy(doing string, arrange func(placement.Point) (arranger.Arrangement, error)) bool {
	if a.ribbon == 0 {
		return false
	}
	at, err := a.ribbonAt()
	if err != nil {
		a.report("reading where the ribbon is", err)
		return false
	}
	arranged, err := arrange(at)
	if err != nil {
		a.report(doing, err)
		return false
	}
	if a.panelOpen.Load() {
		return false
	}
	a.scrolls.Store(arranged.Scrolls)
	a.report("placing the ribbon", a.arrangeWindow(arranged))
	return true
}

// moved records where a drag left the ribbon, putting it back onto a display if the drag left part
// of it off every one (FR-404, FR-406). A move of a panel is not the ribbon's, nor is one of the tab.
func (a *Window) moved() {
	if a.panelOpen.Load() || a.collapsed() {
		return
	}
	window, err := a.position()
	if err != nil {
		a.report("reading where the ribbon was left", err)
		return
	}
	fmt.Fprintf(a.log, "moved: the desktop reports the window at %v (NFR-O-1)\n", window)
	arranged, err := a.service.Moved(a.ribbonFromWindow(window))
	a.report("recording where the ribbon was left", err)
	if err != nil {
		// The placement could not be saved, which raised a notice: fit the ribbon where it stands,
		// its new cell included, rather than leave it wherever the drag let go.
		a.rearrange()
		return
	}
	a.scrolls.Store(arranged.Scrolls)
	// A drop can move the pull out to the ribbon's other side or change the ribbon's size, so the
	// page draws again for the window as it is now cut (measured 2026-10-05).
	a.report("placing the ribbon", a.redrawn(a.arrangeWindow(arranged)))
}

// rearrange fits the ribbon where it stands (FR-104, FR-406).
func (a *Window) rearrange() {
	if a.panelOpen.Load() {
		return
	}
	at, err := a.ribbonAt()
	if err != nil {
		a.report("reading where the ribbon is", err)
		return
	}
	arranged, err := a.service.Rearrange(at)
	if err != nil {
		a.report("fitting the ribbon", err)
		return
	}
	a.scrolls.Store(arranged.Scrolls)
	a.report("placing the ribbon", a.arrangeWindow(arranged))
}

// placeLaunched puts the ribbon where it was last left (FR-405); an unpinned one as its tab.
func (a *Window) placeLaunched() error {
	arranged, err := a.service.Launch()
	if err != nil {
		return err
	}
	fmt.Fprintf(a.log, "launch: placing the ribbon at %v, %v, edge %q (NFR-O-1)\n", arranged.At, arranged.Size, arranged.Edge)
	a.scrolls.Store(arranged.Scrolls)
	return a.arrangeWindow(arranged)
}

// applyAlwaysOnTop keeps the ribbon above other windows where Always on top is on; always while
// unpinned in effect (FR-505, FR-617, FR-619).
func (a *Window) applyAlwaysOnTop() {
	if a.wailsContext() != nil {
		a.unpin.guard.Lock()
		flush := a.unpin.full.Edge != ""
		a.unpin.guard.Unlock()
		a.setOnTop(a.service.Choices().OnTop(flush))
	}
}

// show shows the ribbon; an unpinned one as it stands, its tab while collapsed, which counts as shown
// (FR-618), with the pointer watched for it to open.
func (a *Window) show() {
	if a.wailsContext() == nil {
		return
	}
	a.showWindow()
	a.visible.Store(true)
	a.trackPointer(!a.pinned())
	a.emit(eventRefresh)
}

// hide hides the ribbon, its tab included (FR-618).
func (a *Window) hide() {
	if a.wailsContext() == nil {
		return
	}
	a.hideWindow()
	a.visible.Store(false)
	a.trackPointer(false)
}

func (a *Window) toggle() {
	if a.visible.Load() {
		a.hide()
		return
	}
	a.show()
}

// report writes a failure to the log; nothing when there was none.
func (a *Window) report(doing string, err error) {
	if err != nil {
		fmt.Fprintf(a.log, "%s: %v\n", doing, err)
	}
}
