//go:build linux || darwin

package desktop

import (
	"io"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
)

// quietFor is how long a test waits to be sure no event comes: well past the settling time.
const quietFor = 3 * moveSettle

// watchedDesktop answers a desktop watching a fresh test window, with both cleaned up afterwards.
func watchedDesktop(t *testing.T) (*Desktop, Window) {
	t.Helper()
	ribbon := newTestWindow(testApp.Name)
	d := New(testApp, func() []menus.Item { return nil }, io.Discard)
	d.guard.Lock()
	d.ribbon = ribbon
	d.guard.Unlock()
	t.Cleanup(func() {
		d.Stop()
		closeTestWindow(ribbon)
	})
	return d, ribbon
}

// FR-404: arriving where it was placed is not a move; standing somewhere else is, once it settles.
func TestOnlyAMoveTheRibbonWasNotPlacedIntoEndsAsAMove(t *testing.T) {
	d, ribbon := watchedDesktop(t)
	placedTo := placement.Point{X: 200, Y: 150}
	if err := Place(ribbon, placedTo, testSize); err != nil {
		t.Fatal(err)
	}
	d.moved(placedTo)
	select {
	case event := <-d.Events():
		t.Fatalf("a placement was reported as event %d", event.Kind)
	case <-time.After(quietFor):
	}
	d.moved(placement.Point{X: 260, Y: 150})
	d.moved(placement.Point{X: 320, Y: 150})
	select {
	case event := <-d.Events():
		if event.Kind != EventMoveEnded {
			t.Errorf("got event %d", event.Kind)
		}
	case <-time.After(quietFor):
		t.Fatal("a drag was not reported as ended")
	}
	select {
	case <-d.Events():
		t.Error("one drag was reported twice")
	case <-time.After(quietFor):
	}
}

// A window that stands somewhere else for a moment on its way to where it was placed has not been
// moved (measured 2026-09-28: a ribbon growing into a panel was clamped by the window manager before
// its move arrived).
func TestAPassingPositionOnTheWayToAPlacementIsNotAMove(t *testing.T) {
	d, ribbon := watchedDesktop(t)
	placedTo := placement.Point{X: 200, Y: 150}
	if err := Place(ribbon, placedTo, testSize); err != nil {
		t.Fatal(err)
	}
	d.moved(placement.Point{X: 880, Y: 152})
	d.moved(placedTo)
	select {
	case event := <-d.Events():
		t.Errorf("a passing position was reported as event %d", event.Kind)
	case <-time.After(quietFor):
	}
}

// FR-108: the item chosen from the ribbon's menu reaches the desktop as its action.
func TestTheChosenMenuItemIsReported(t *testing.T) {
	d, _ := watchedDesktop(t)
	items := []menus.Item{
		{Action: menus.Settings},
		{Label: "Help", Children: []menus.Item{{Action: menus.About}}},
		{Action: menus.Exit},
	}
	shown.Lock()
	shown.desktop, shown.items = d, items
	shown.Unlock()
	desktopMenuChosen(1)
	if event := <-d.Events(); event.Kind != EventMenu || event.Action != menus.About {
		t.Errorf("got %+v", event)
	}
}

// An event after Stop is dropped, never sent on the closed channel.
func TestAnEventAfterStopIsDropped(t *testing.T) {
	d, _ := watchedDesktop(t)
	d.Stop()
	d.send(Event{Kind: EventDisplayChanged})
	d.moved(placement.Point{X: 1, Y: 1})
	if _, open := <-d.Events(); open {
		t.Error("the events are still open after Stop")
	}
}
