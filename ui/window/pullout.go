package window

// The pull out beside the ribbon (FR-901 to FR-913). The pull out shares the ribbon's window: while it
// shows, the window is the ribbon and the pull out together; the page lays the two out inside it.
// What the pull out draws is the application's.
// Every placement is decided for the ribbon alone; this file turns it into the window's and back.

import (
	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/domain/placement"
)

// windowOf answers the window's place and size for the ribbon arranged as full, shown in full: the
// ribbon with its pull out where one shows. offset is where the ribbon's corner lies inside the window.
func windowOf(full arranger.Arrangement) (at placement.Point, size placement.Size, offset placement.Point) {
	ribbon := placement.Rect{Left: full.At.X, Top: full.At.Y, Right: full.At.X + full.Size.Width, Bottom: full.At.Y + full.Size.Height}
	whole := ribbon
	if full.PullOut != (placement.Rect{}) {
		whole = placement.Rect{
			Left: min(ribbon.Left, full.PullOut.Left), Top: min(ribbon.Top, full.PullOut.Top),
			Right: max(ribbon.Right, full.PullOut.Right), Bottom: max(ribbon.Bottom, full.PullOut.Bottom),
		}
	}
	at = placement.Point{X: whole.Left, Y: whole.Top}
	return at, placement.Size{Width: whole.Width(), Height: whole.Height()}, placement.Point{X: ribbon.Left - whole.Left, Y: ribbon.Top - whole.Top}
}

// ribbonFromWindow answers where the ribbon stands for a window at at showing the full ribbon,
// which a drag moves as one.
func (a *Window) ribbonFromWindow(at placement.Point) placement.Point {
	a.unpin.guard.Lock()
	_, _, offset := windowOf(a.unpin.full)
	a.unpin.guard.Unlock()
	return placement.Point{X: at.X + offset.X, Y: at.Y + offset.Y}
}

// TogglePullOut opens a vertical ribbon's pull out when closed and closes it when open (FR-903), as
// its handle is clicked. The window changes size, so the page is told to draw it again: without that
// it kept the closed layout in a window grown for the pull out (Oliver, 2026-09-29).
func (a *Window) TogglePullOut() error {
	return a.redrawn(a.refitted(a.service.SetPullOut(!a.service.PullOut())))
}

// placeShaped cuts the window to parts (FR-913), then puts it at at, size across. The cut comes
// first, in the new window's pixels, so a window growing for the pull out never shows the bands it is
// cut from. A failed cut is written to the log and the window is placed all the same, as a rectangle.
func (a *Window) placeShaped(at placement.Point, size placement.Size, parts []placement.Rect) error {
	a.report("cutting the window to the ribbon and its pull out", a.shape(parts))
	a.lastPlaced.note(at)
	return a.place(at, size)
}

// placeWhole puts the window at at, size across, keeping all of it: the tab and a panel.
func (a *Window) placeWhole(at placement.Point, size placement.Size) error {
	return a.placeShaped(at, size, placement.Shape(size, placement.Rect{}, placement.Rect{}, false))
}

// redrawn tells the page to take a fresh snapshot, then answers err.
func (a *Window) redrawn(err error) error {
	a.emit(eventRefresh)
	return err
}

// pullOutLayout answers where the page draws the ribbon and its pull out inside the window, in the
// window's pixels: the pull out shown only while the full ribbon is, never with the tab or a panel
// (FR-910). An opening ribbon counts as shown while the page draws it, as collapsed agrees, since the
// window grows round what was drawn then (FR-615).
func (a *Window) pullOutLayout() (side placement.Edge, ribbon, pullOut placement.Rect, shown bool) {
	a.unpin.guard.Lock()
	full := a.unpin.full
	open := (a.unpin.shownOpen || a.unpin.drawing) && !a.panelOpen.Load()
	a.unpin.guard.Unlock()
	at, _, offset := windowOf(full)
	ribbon = placement.Rect{Left: offset.X, Top: offset.Y, Right: offset.X + full.Size.Width, Bottom: offset.Y + full.Size.Height}
	if !open || full.PullOut == (placement.Rect{}) {
		return full.PullOutSide, ribbon, placement.Rect{}, false
	}
	pullOut = placement.Rect{Left: full.PullOut.Left - at.X, Top: full.PullOut.Top - at.Y, Right: full.PullOut.Right - at.X, Bottom: full.PullOut.Bottom - at.Y}
	return full.PullOutSide, ribbon, pullOut, true
}
