//go:build linux || darwin

package desktop

import (
	"fmt"
	"io"
	"runtime/cgo"
	"sync"
	"time"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/identity"
	"github.com/oernster/ribbonkit/domain/placement"
)

// moveSettle is how long the ribbon must stand still before a move counts as ended. The window
// manager (AppKit's own drag on macOS) carries a move through without saying when the button is
// let go, so the end is the moment the position stops changing.
const moveSettle = 300 * time.Millisecond

// Desktop hears the desktop on Linux and macOS: the tray icon and its menu, the end of the ribbon's
// moves, changes of display and jumps of the clock, plus the ribbon's own menu. The toolkit's loop
// is not running when Start runs, so everything that needs the ribbon's window waits for Watch,
// which runs once that window exists. The tray starts with Start: on Linux it talks to the session
// bus rather than GTK; on macOS it is set up once AppKit's loop runs. Each platform supplies
// startTray, watchWindow, popUp and the ribbon's operations in files of its own.
type Desktop struct {
	app    identity.App
	menu   func() []menus.Item
	events chan Event
	log    io.Writer
	stop   chan struct{}
	icon   []byte
	tray   *tray

	guard  sync.Mutex
	closed bool
	ribbon Window
	lastAt placement.Point
	settle *time.Timer

	pointer pointerTracker

	started sync.Once
	watched sync.Once
	stopped sync.Once
}

// SetTabFrame does nothing off Windows: GTK and AppKit take the tab's 8 as asked (measured
// 2026-09-28, REQUIREMENTS section 2.3), so no style holds the window wider.
func SetTabFrame(Window, bool) error { return nil }

// New answers a desktop for app whose tray menu is menu, reporting failures to log.
func New(app identity.App, menu func() []menus.Item, log io.Writer) *Desktop {
	return &Desktop{app: app, menu: menu, events: make(chan Event, eventBuffer), log: log, stop: make(chan struct{})}
}

// Events yields what happened. The channel is closed when the desktop stops.
func (d *Desktop) Events() <-chan Event { return d.events }

// UseIcon gives the tray icon its image, a PNG. It must come before Start.
func (d *Desktop) UseIcon(png []byte) { d.icon = png }

// Start begins watching the clock and puts the icon in the tray, returning once the icon is there or
// with the reason it is not.
func (d *Desktop) Start() error {
	var err error
	d.started.Do(func() {
		go watchClock(d.stop, func() { d.send(Event{Kind: EventTimeChanged}) })
		d.tray, err = startTray(d, d.icon)
	})
	return err
}

// Stop takes the icon out of the tray, ends the watching and closes the events.
func (d *Desktop) Stop() {
	d.stopped.Do(func() {
		if d.tray != nil {
			d.tray.stop()
		}
		close(d.stop)
		d.guard.Lock()
		defer d.guard.Unlock()
		d.closed = true
		if d.settle != nil {
			d.settle.Stop()
		}
		close(d.events)
	})
}

// Watch names the ribbon's window, so the end of its moves is reported, then hears the changes of
// display. It runs once the toolkit's loop is running.
func (d *Desktop) Watch(ribbon Window) {
	d.watched.Do(func() {
		d.guard.Lock()
		d.ribbon = ribbon
		d.guard.Unlock()
		if err := watchWindow(ribbon, cgo.NewHandle(d)); err != nil {
			fmt.Fprintf(d.log, "desktop: watching the ribbon: %v\n", err)
		}
	})
}

// ShowMenu shows items as a native menu at the pointer, from any goroutine: the ribbon's right-click
// menu (FR-108). The choice arrives as an EventMenu. Before the ribbon is found there is nothing to
// show it over.
func (d *Desktop) ShowMenu(items []menus.Item) {
	d.guard.Lock()
	ribbon := d.ribbon
	d.guard.Unlock()
	if ribbon == 0 {
		return
	}
	if err := popUp(ribbon, d, items); err != nil {
		fmt.Fprintf(d.log, "desktop: showing the menu: %v\n", err)
	}
}

// moved hears the ribbon standing at at. A position the ribbon was placed at is not a move; it also
// cancels any wait already begun, since on the way there the window manager may stand it somewhere
// else for a moment, as when it grows before it moves. Any other position starts the wait for it
// to settle, the wait starting over while it keeps moving.
func (d *Desktop) moved(at placement.Point) {
	d.guard.Lock()
	defer d.guard.Unlock()
	placedTo, placedOK := placedAt(d.ribbon)
	if d.closed || at == d.lastAt {
		return
	}
	d.lastAt = at
	if d.settle != nil {
		d.settle.Stop()
	}
	if placedOK && placedTo == at {
		return
	}
	d.settle = time.AfterFunc(moveSettle, func() { d.send(Event{Kind: EventMoveEnded}) })
}

// send hands event on without waiting: one nobody is reading is dropped and said so, since the
// desktop calls in on the toolkit's own thread, which must never block.
func (d *Desktop) send(event Event) {
	d.guard.Lock()
	defer d.guard.Unlock()
	if d.closed {
		return
	}
	select {
	case d.events <- event:
	default:
		fmt.Fprintf(d.log, "desktop: event %d dropped, nothing is reading\n", event.Kind)
	}
}
