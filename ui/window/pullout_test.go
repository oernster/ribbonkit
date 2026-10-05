package window

import (
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/shell"
	"github.com/oernster/ribbonkit/domain/placement"
)

// withPullOut is testArrange, a ribbon at (10, 20) 300 by 90, with a 480 by 240 pull out below it
// centred on it: the window runs from x -80 to 400 and y 20 to 350, the ribbon's corner 90 in from
// its left.
var withPullOut = func() arranger.Arrangement {
	arranged := testArrange
	arranged.PullOutSide = placement.Bottom
	arranged.PullOut = placement.Rect{Left: -80, Top: 110, Right: 400, Bottom: 350}
	return arranged
}()

// FR-902, FR-909: the window is the ribbon with its pull out; the ribbon's corner sits inside it.
func TestTheWindowHoldsTheRibbonAndItsPullOut(t *testing.T) {
	t.Parallel()
	at, size, offset := windowOf(withPullOut)
	if at != (placement.Point{X: -80, Y: 20}) || size != (placement.Size{Width: 480, Height: 330}) || offset != (placement.Point{X: 90, Y: 0}) {
		t.Errorf("got %+v %+v %+v", at, size, offset)
	}
	if at, size, offset := windowOf(testArrange); at != testArrange.At || size != testArrange.Size || offset != (placement.Point{}) {
		t.Errorf("no pull out: got %+v %+v %+v", at, size, offset)
	}
}

// FR-909: a drag moves the window; the ribbon's own place is read back through its corner's offset,
// so the drop is decided for the ribbon, never the window.
func TestADragOfThePullOutMovesTheRibbonToo(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.arrangement = withPullOut
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	if got := lastPlaced(t, seen); got.At != (placement.Point{X: -80, Y: 20}) || got.Size != (placement.Size{Width: 480, Height: 330}) {
		t.Fatalf("placed %+v, want the window with its pull out", got)
	}
	seen.ribbonAt = placement.Point{X: 20, Y: 500}
	app.handleSafely(shell.Event{Kind: shell.EventMoveEnded})
	if at := service.at[len(service.at)-1]; at != (placement.Point{X: 110, Y: 500}) {
		t.Errorf("the drop was read as %+v, want the ribbon's corner 90 in", at)
	}
}

// FR-909, FR-910: a drop that changes the side the pull out goes on tells the page to draw again, so
// it lays the two out as the window is now cut. Measured 2026-10-05: dragged onto a display at 250
// percent and back, the window was cut for the map on the ribbon's left while the page still drew
// it on the right.
func TestADropThatMovesThePullOutTellsThePage(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.arrangement = withPullOut
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	above := withPullOut
	above.PullOutSide = placement.Top
	above.PullOut = placement.Rect{Left: -80, Top: -220, Right: 400, Bottom: 20}
	service.arrangement = above
	seen.events = nil
	app.handleSafely(shell.Event{Kind: shell.EventMoveEnded})
	if !seen.sawEvent(eventRefresh) {
		t.Errorf("the drop sent %v, want the page told to draw again", seen.events)
	}
	if got := app.shown(); got.PullOutSide != placement.Top {
		t.Errorf("the page would be told %+v, want the pull out on top", got)
	}
}

// FR-407, FR-909: a new pixel ratio, as when a drag carries the window onto a display at other
// scaling, refits the ribbon, which can move the pull out to its other side; the page is told to draw
// again. Measured 2026-10-05: crossing onto a display at 250 percent mid-drag drew the old layout in
// the window cut for the new one until the drop.
func TestANewPixelRatioTellsThePage(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.arrangement = withPullOut
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	above := withPullOut
	above.PullOutSide = placement.Top
	above.PullOut = placement.Rect{Left: -80, Top: -220, Right: 400, Bottom: 20}
	service.arrangement = above
	seen.events = nil
	if err := app.SetPixelRatio(2.5); err != nil {
		t.Fatal(err)
	}
	if !seen.sawEvent(eventRefresh) || app.shown().PullOutSide != placement.Top {
		t.Errorf("a new ratio sent %v with the pull out at %q, want the page told to draw it on top", seen.events, app.shown().PullOutSide)
	}
}

// FR-910: the page is told where to draw the pull out only while the full ribbon shows, never with a
// tab.
func TestThePullOutHidesWithTheTab(t *testing.T) {
	t.Parallel()
	app, service, _ := unpinnedApp(t)
	service.arrangement = withPullOut
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	if shown := app.shown(); shown.PullOutShown || !app.collapsed() {
		t.Errorf("collapsed: %+v", shown)
	}
	service.choices.Pinned = true
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	got := app.shown()
	if !got.PullOutShown || got.PullOutSide != placement.Bottom || got.PullOut != (Box{X: 0, Y: 90, Width: 480, Height: 240}) || got.Ribbon != (Box{X: 90, Y: 0, Width: 300, Height: 90}) {
		t.Errorf("shown: %+v", got)
	}
}

// FR-615, FR-903: the page is told where the ribbon and its pull out go in its own units, the window
// pixels divided by the pixels to each unit that windows are sized with, so it draws them at their
// size even while the window is still the tab. A ratio the service refused changes nothing.
func TestThePullOutsPartsReachThePageInItsOwnUnits(t *testing.T) {
	t.Parallel()
	app, service, _ := unpinnedApp(t)
	service.arrangement = withPullOut
	service.choices.Pinned = true
	const ratio = 1.25
	if err := app.SetPixelRatio(ratio); err != nil {
		t.Fatal(err)
	}
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	// The stand-in desktop answers the page's ratio as it is.
	perDIP := float64(ratio)
	want := Box{X: 90 / perDIP, Y: 0, Width: 300 / perDIP, Height: 90 / perDIP}
	if got := app.shown(); got.Ribbon != want || got.PullOut.Width != 480/perDIP {
		t.Errorf("at %v pixels to a unit the page was told %+v, want the ribbon at %+v", perDIP, got, want)
	}
	service.changeErr = errPlanted
	_ = app.SetPixelRatio(2 * ratio)
	if got := app.shown().Ribbon; got != want {
		t.Errorf("a refused ratio moved the ribbon to %+v", got)
	}
}

// FR-615, FR-910: an opening ribbon is drawn before the window grows, so the page is told of its pull
// out while it draws, not only once grown. Measured 2026-09-29: told of none, the page drew the ribbon
// alone and the window then grew round a blank pull out.
func TestAnOpeningRibbonIsDrawnWithItsPullOut(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.choices.Pinned = false
	service.arrangement = withPullOut
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	app.handleSafely(shell.Event{Kind: shell.EventPointerArrived})
	fire(t, seen)
	if seen.drawPending == nil {
		t.Fatal("opening did not wait for the page")
	}
	got := app.shown()
	if got.Collapsed || !got.PullOutShown || got.PullOut != (Box{X: 0, Y: 90, Width: 480, Height: 240}) {
		t.Errorf("while drawing the page was told %+v; want the ribbon with its pull out", got)
	}
}

// FR-913: every placing of the window is cut first, to the ribbon and its pull out while the pull out
// shows and to the whole window otherwise: no pull out, the tab, a panel. A cut that fails still
// places the window.
func TestTheShapeFollowsEveryRefit(t *testing.T) {
	t.Parallel()
	app, service, seen, log := newTestApp(t)
	service.arrangement = withPullOut
	app.place = func(at placement.Point, size placement.Size) error {
		if len(seen.shapes) != len(seen.placed)+1 {
			t.Errorf("placed after %d cuts, want the cut first", len(seen.shapes))
		}
		seen.placed = append(seen.placed, arranger.Arrangement{At: at, Size: size})
		return nil
	}
	lastShape := func() []placement.Rect { return seen.shapes[len(seen.shapes)-1] }
	whole := func(size placement.Size) []placement.Rect {
		return []placement.Rect{{Right: size.Width, Bottom: size.Height}}
	}
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	ribbon := placement.Rect{Left: 90, Top: 0, Right: 390, Bottom: 90}
	pullOut := placement.Rect{Left: 0, Top: 90, Right: 480, Bottom: 330}
	if got := lastShape(); len(got) != 2 || got[0] != ribbon || got[1] != pullOut {
		t.Errorf("with the pull out: cut to %+v", got)
	}
	service.arrangement = testArrange
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	if got := lastShape(); !slices.Equal(got, whole(testArrange.Size)) {
		t.Errorf("no pull out: cut to %+v", got)
	}
	if err := app.OpenPanel(openAtAbout); err != nil {
		t.Fatal(err)
	}
	if got := lastShape(); !slices.Equal(got, whole(lastPlaced(t, seen).Size)) {
		t.Errorf("panel: cut to %+v", got)
	}
	app.shape = func([]placement.Rect) error { return errPlanted }
	placedBefore := len(seen.placed)
	app.place = func(at placement.Point, size placement.Size) error {
		seen.placed = append(seen.placed, arranger.Arrangement{At: at, Size: size})
		return nil
	}
	if err := app.ClosePanel(); err != nil || len(seen.placed) == placedBefore {
		t.Errorf("a failed cut stopped the placing (%v)", err)
	}
	if !strings.Contains(log.String(), "cutting the window") {
		t.Errorf("the failed cut was not logged: %q", log.String())
	}
}

// FR-913: the tab is never cut; it keeps all of itself.
func TestTheTabIsNeverCut(t *testing.T) {
	t.Parallel()
	app, service, seen := unpinnedApp(t)
	service.arrangement = withPullOut
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	if got := seen.shapes[len(seen.shapes)-1]; !app.collapsed() || !slices.Equal(got, []placement.Rect{{Right: lastPlaced(t, seen).Size.Width, Bottom: lastPlaced(t, seen).Size.Height}}) {
		t.Errorf("tab cut to %+v", got)
	}
}

// FR-903: the handle flips the pull out, then refits the window and has the page draw it again.
func TestTheHandleFlipsThePullOut(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	if err := app.TogglePullOut(); err != nil {
		t.Fatal(err)
	}
	if !seen.sawEvent(eventRefresh) || !slices.Contains(service.calls, "Rearrange") {
		t.Errorf("the handle changed the window but it was not refitted (%v) or the page not told", service.calls)
	}
	if !service.pullOut {
		t.Error("the first click left the pull out closed")
	}
	if err := app.TogglePullOut(); err != nil || service.pullOut {
		t.Errorf("the second click left the pull out %v (%v)", service.pullOut, err)
	}
}
