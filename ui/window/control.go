package window

import (
	"io/fs"
	"math"
	"os"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
)

// Control is the application's hold on its window. It is kept in a named field and never embedded,
// so nothing here is bound to the page.
type Control struct {
	window *Window
}

// Shown is how the window shows the ribbon now, which the application's snapshot carries to the page.
type Shown struct {
	// Collapsed is true while the window is an unpinned ribbon's tab (FR-614).
	Collapsed bool
	// Scrolls is true when the cells need more room than the work area offers (FR-106).
	Scrolls bool
	// DragThreshold is how far the pointer moves before a press becomes a drag.
	DragThreshold placement.Size
	// PullOutSide is where the pull out and its handle go, empty while there is none; PullOutShown is
	// whether it is drawn now, in PullOut, beside the ribbon in Ribbon (FR-901 to FR-910).
	PullOutSide  placement.Edge
	PullOutShown bool
	Ribbon       Box
	PullOut      Box
}

// Refitted fits the ribbon after a change of the application's own, then answers err (FR-707).
func (c *Control) Refitted(err error) error { return c.window.refitted(err) }

// ContentChanged fits the ribbon to what it now holds where it stands (FR-104, FR-105).
func (c *Control) ContentChanged() { c.window.contentChanged() }

// Redraw tells the page to take a fresh snapshot.
func (c *Control) Redraw() { c.window.emit(eventRefresh) }

// Redrawn tells the page to take a fresh snapshot, then answers err.
func (c *Control) Redrawn(err error) error { return c.window.redrawn(err) }

// Report writes a failure to the log; nothing when there was none.
func (c *Control) Report(doing string, err error) { c.window.report(doing, err) }

// ShowPanel shows the ribbon, then has the page open panel, a word the page knows it by.
func (c *Control) ShowPanel(panel string) {
	c.window.show()
	c.window.emit(eventOpenPanel, panel)
}

// PageMeasured records that the page's widest text has been applied to the window, which the
// launched ribbon's first showing waits for (launch_show.go).
func (c *Control) PageMeasured() { c.window.pageMeasured() }

// Shown answers how the window shows the ribbon now.
func (c *Control) Shown() Shown { return c.window.shown() }

// Run runs the window, binding bound to the page and serving it from assets, with the web view's
// data kept in dataDir (run.go).
func (c *Control) Run(bound any, assets fs.FS, dataDir string) error {
	return c.window.run(bound, assets, dataDir)
}

// ExitWhen ends the application once a signal arrives on signals (quit_signal.go).
func (c *Control) ExitWhen(signals <-chan os.Signal) { c.window.exitWhen(signals) }

// TrayStarted records that the tray icon is up, so closing the ribbon hides it rather than ending
// the application (FR-507).
func (c *Control) TrayStarted() { c.window.trayUp.Store(true) }

// Offered answers items as a menu or Settings offers them: each Position item greyed where pressing it
// would leave the ribbon where it stands (FR-408, FR-412). The tray's menu and the Settings choices
// pass through it, as the ribbon's own menu does.
func (c *Control) Offered(items []menus.Item) []menus.Item { return c.window.offered(items) }

// Visible answers whether the ribbon is shown, which the tray menu offers to change.
func (c *Control) Visible() bool { return c.window.visible.Load() }

// shown answers how the window shows the ribbon now, the pull out's boxes in the page's units.
func (a *Window) shown() Shown {
	side, ribbon, pullOut, drawn := a.pullOutLayout()
	perDIP := math.Float64frombits(a.pixelsPerDIP.Load())
	return Shown{
		Collapsed: a.collapsed(), Scrolls: a.scrolls.Load(), DragThreshold: a.dragThreshold(),
		PullOutSide: side, PullOutShown: drawn, Ribbon: boxOf(ribbon, perDIP), PullOut: boxOf(pullOut, perDIP),
	}
}

// boxOf answers r, a rectangle in window pixels, in the page's units: divided by perDIP, the window
// pixels to each of them. The page then draws the box as it is, whatever size the window has at that
// moment; an opening ribbon is drawn while the window is still its tab (FR-615). Before the page has
// reported its ratio perDIP is zero and r goes as it is, one pixel to a unit.
func boxOf(r placement.Rect, perDIP float64) Box {
	if perDIP == 0 {
		perDIP = 1
	}
	return Box{
		X: float64(r.Left) / perDIP, Y: float64(r.Top) / perDIP,
		Width: float64(r.Width()) / perDIP, Height: float64(r.Height()) / perDIP,
	}
}
