package window

import (
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/shell"
	"github.com/oernster/ribbonkit/domain/hover"
	"github.com/oernster/ribbonkit/domain/placement"
)

// unpinnedApp answers a facade whose ribbon is unpinned, launched and shown: its tab.
func unpinnedApp(t *testing.T) (*Window, *scriptedService, *window) {
	t.Helper()
	app, service, seen, _ := newTestApp(t)
	service.choices.Pinned = false
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	return app, service, seen
}

// fire makes the pending timer fall due, as time.AfterFunc would once its wait was over.
func fire(t *testing.T, seen *window) {
	t.Helper()
	do := seen.pending
	if do == nil {
		t.Fatal("nothing is pending")
	}
	seen.pending = nil
	seen.now = seen.now.Add(seen.waited)
	do()
}

func lastPlaced(t *testing.T, seen *window) arranger.Arrangement {
	t.Helper()
	if len(seen.placed) == 0 {
		t.Fatal("nothing was placed")
	}
	return seen.placed[len(seen.placed)-1]
}

// FR-614: launched unpinned, the window is the tab framed for it; the page is told to draw it.
func TestLaunchingUnpinnedShowsTheTab(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness || got.At.X != testArrange.At.X+testArrange.Size.Width-placement.TabThickness {
		t.Errorf("placed %+v, want the tab", got)
	}
	if len(seen.tabFrames) == 0 || !seen.tabFrames[len(seen.tabFrames)-1] {
		t.Errorf("frames %v, want the tab's last", seen.tabFrames)
	}
	if !app.collapsed() {
		t.Error("the snapshot does not say the ribbon is collapsed")
	}
	if len(seen.watching) == 0 || !seen.watching[len(seen.watching)-1] {
		t.Error("the pointer is not watched while the tab is shown")
	}
}

// FR-615, FR-616: resting on the tab opens the full ribbon; leaving collapses it a second later.
func TestTheTabOpensAfterTheRestAndCollapsesOnceAway(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	app.handleSafely(shell.Event{Kind: shell.EventPointerArrived})
	if seen.waited != hover.Rest {
		t.Fatalf("waiting %v, want the rest", seen.waited)
	}
	fire(t, seen)
	if app.collapsed() || !seen.sawEvent(eventRefresh) {
		t.Error("the page was not told the ribbon opened")
	}
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness {
		t.Errorf("grew to %+v before the page had drawn", got)
	}
	app.RibbonDrawn()
	if got := lastPlaced(t, seen); got != (arranger.Arrangement{At: testArrange.At, Size: testArrange.Size}) {
		t.Errorf("opened at %+v, want the full ribbon", got)
	}
	app.handleSafely(shell.Event{Kind: shell.EventPointerLeft})
	if seen.waited != hover.Away {
		t.Fatalf("waiting %v, want the second away", seen.waited)
	}
	fire(t, seen)
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness {
		t.Errorf("collapsed to %+v, want the tab", got)
	}
}

// FR-616: a panel and the ribbon's own menu each hold it open; it collapses once both have ended.
func TestAPanelAndTheMenuHoldTheRibbonOpen(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	app.ShowContextMenu()
	if err := app.OpenPanel(openAtAbout); err != nil {
		t.Fatal(err)
	}
	app.handleSafely(shell.Event{Kind: shell.EventPointerLeft})
	app.handleSafely(shell.Event{Kind: shell.EventMenuClosed})
	if seen.pending != nil {
		t.Fatal("the ribbon would collapse while its panel stands")
	}
	if err := app.ClosePanel(); err != nil {
		t.Fatal(err)
	}
	if got := lastPlaced(t, seen); got.Size != testArrange.Size {
		t.Errorf("the panel closed onto %+v, want the full ribbon", got)
	}
	if seen.pending == nil || seen.waited != hover.Away {
		t.Errorf("after the panel closed: waiting %v, want the second away", seen.waited)
	}
}

// FR-613, FR-617: pinning shows the ribbon in full and stops watching the pointer; unpinning keeps it
// on top whatever Always on top holds.
func TestPinningAndUnpinning(t *testing.T) {
	t.Parallel()
	app, service, seen := unpinnedApp(t)
	app.act(menus.Pin)
	app.RibbonDrawn()
	if !service.choices.Pinned || lastPlaced(t, seen).Size != testArrange.Size {
		t.Errorf("pinning left pinned %v, placed %+v", service.choices.Pinned, lastPlaced(t, seen))
	}
	if seen.tabFrames[len(seen.tabFrames)-1] {
		t.Error("pinned, the ribbon still wears the tab's frame")
	}
	if seen.watching[len(seen.watching)-1] || seen.onTop[len(seen.onTop)-1] {
		t.Error("pinned, the pointer is still watched or the ribbon kept on top")
	}
	app.handleSafely(shell.Event{Kind: shell.EventPointerLeft})
	if seen.pending != nil {
		t.Error("a pinned ribbon heard the pointer")
	}
	app.act(menus.Pin)
	if service.choices.Pinned || !seen.onTop[len(seen.onTop)-1] || seen.waited != hover.Away {
		t.Errorf("unpinning: pinned %v, on top %v, waiting %v", service.choices.Pinned, seen.onTop, seen.waited)
	}
	if !seen.tabFrames[len(seen.tabFrames)-1] || lastPlaced(t, seen).Size != testArrange.Size {
		t.Errorf("unpinned in full: frames %v, placed %+v; want the tab's frame on the full ribbon", seen.tabFrames, lastPlaced(t, seen))
	}
}

// openedToDrawing rests the pointer on the tab until the ribbon opens, leaving the page drawing it.
func openedToDrawing(t *testing.T) (*Window, *window) {
	t.Helper()
	app, _, seen := unpinnedApp(t)
	app.handleSafely(shell.Event{Kind: shell.EventPointerArrived})
	fire(t, seen)
	if seen.drawPending == nil {
		t.Fatal("opening did not wait for the page")
	}
	return app, seen
}

// Opening grows the window once the page has drawn, keeping the tab's frame, which Windows painted a
// caption over when Wails' came back (measured 2026-09-29). A second word from the page does nothing.
func TestOpeningGrowsOnceThePageHasDrawn(t *testing.T) {
	t.Parallel()
	app, seen := openedToDrawing(t)
	app.RibbonDrawn()
	if seen.drawPending != nil || lastPlaced(t, seen).Size != testArrange.Size {
		t.Errorf("drawn: fallback pending %v, placed %+v", seen.drawPending != nil, lastPlaced(t, seen))
	}
	if !seen.tabFrames[len(seen.tabFrames)-1] {
		t.Error("the unpinned ribbon opened in Wails' frame")
	}
	placed := len(seen.placed)
	app.RibbonDrawn()
	if len(seen.placed) != placed {
		t.Error("a second word from the page placed the ribbon again")
	}
}

// A page that never says it has drawn still has its ribbon opened once drawWait has passed.
func TestOpeningGrowsWhenThePageNeverAnswers(t *testing.T) {
	t.Parallel()
	app, seen := openedToDrawing(t)
	seen.drawPending()
	if lastPlaced(t, seen).Size != testArrange.Size || app.collapsed() {
		t.Errorf("after the wait: placed %+v", lastPlaced(t, seen))
	}
}

// Leaving before the page has drawn collapses the ribbon back; the page's late word changes nothing.
func TestLeavingWhileThePageDrawsCollapsesAgain(t *testing.T) {
	t.Parallel()
	app, seen := openedToDrawing(t)
	app.handleSafely(shell.Event{Kind: shell.EventPointerLeft})
	fire(t, seen)
	if seen.drawPending != nil || !app.collapsed() {
		t.Errorf("fallback pending %v, collapsed %v", seen.drawPending != nil, app.collapsed())
	}
	app.RibbonDrawn()
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness {
		t.Errorf("the late word placed %+v, want the tab kept", got)
	}
}

// A panel opened while the page draws is the window; the page's word does not put the ribbon over it.
func TestAPanelOpenedWhileThePageDrawsStays(t *testing.T) {
	t.Parallel()
	app, seen := openedToDrawing(t)
	if err := app.OpenPanel(openAtAbout); err != nil {
		t.Fatal(err)
	}
	placed := len(seen.placed)
	app.RibbonDrawn()
	if len(seen.placed) != placed {
		t.Errorf("placed %+v over the panel", lastPlaced(t, seen))
	}
}

// FR-618: a collapsed ribbon counts as shown, so a second launch hides it, tab included.
func TestASecondLaunchHidesACollapsedRibbon(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	app.secondInstance()
	if seen.hidden != 1 || seen.watching[len(seen.watching)-1] {
		t.Errorf("hidden %d, watching %v; want hidden with the pointer unwatched", seen.hidden, seen.watching)
	}
}

// FR-618: the tray menu is built from whether the ribbon is visible (main.go hands TrayMenu the
// facade's visible), so a collapsed ribbon counts as shown and the tray offers to hide it; a second
// launch hides it, when the tray offers to show it; a third brings the tab back, still collapsed.
func TestTheTrayMenuTreatsACollapsedRibbonAsShown(t *testing.T) {
	t.Parallel()
	app, _, _ := unpinnedApp(t)
	if !app.collapsed() {
		t.Fatal("the unpinned ribbon did not start as its tab")
	}
	if !app.visible.Load() {
		t.Error("a collapsed ribbon is not counted as shown, so the tray would offer Show ribbon")
	}
	app.secondInstance()
	if app.visible.Load() {
		t.Error("a second launch left the collapsed ribbon counted as shown")
	}
	app.secondInstance()
	if !app.visible.Load() || !app.collapsed() {
		t.Errorf("a third launch: visible %v, collapsed %v; want the tab back", app.visible.Load(), app.collapsed())
	}
}

// A panic while opening or collapsing, on the timer's own goroutine, is logged, not fatal.
func TestAFailureWhileOpeningIsLogged(t *testing.T) {
	t.Parallel()
	app, _, seen, log := newTestApp(t)
	app.service.(*scriptedService).choices.Pinned = false
	app.handleSafely(shell.Event{Kind: shell.EventPointerArrived})
	app.handleSafely(shell.Event{Kind: shell.EventPointerLeft})
	app.handleSafely(shell.Event{Kind: shell.EventPointerArrived})
	app.place = func(placement.Point, placement.Size) error { panic("planted") }
	app.unpin.shownOpen = false
	fire(t, seen)
	seen.drawPending()
	if !strings.Contains(log.String(), "recovered from planted") {
		t.Errorf("log %q", log.String())
	}
}
