// Package arranger places and sizes a ribbon on the displays: where it launches, where a drag or an
// edge puts it, how large its cells make it at the scale it is drawn at, the tab it collapses to and
// the pull out it may hold beside it. What the ribbon shows is the application's; the arranger asks for it
// through Host, so every ribbon is arranged by the same rules.
//
// FR numbers are TimeRibbon's REQUIREMENTS.md, where each rule was first specified.
package arranger

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// ErrNoMonitors is answered when the system reports no display at all.
var ErrNoMonitors = errors.New("no display is reported")

// ErrNotAgainstAnEdge is answered when the tab of a ribbon flush against no edge is asked for; such a
// ribbon never collapses (FR-619).
var ErrNotAgainstAnEdge = errors.New("the ribbon stands against no edge, so it has no tab")

// ErrUnusableScale is answered when the page reports a scale that is not a positive number.
var ErrUnusableScale = errors.New("a scale must be a positive number")

// Content is what the application draws in the ribbon, as far as its size depends on it; every
// length is in DIP at the ribbon's own size, before its scale.
type Content struct {
	// Cell is one cell's size, its width running along a horizontal ribbon.
	Cell placement.Size
	// Cells counts every cell drawn; at least one.
	Cells int
	// Padding surrounds the cells on every side.
	Padding int
	// Lane is depth added across the ribbon beside its cells, such as the lane a pull out's handle
	// stands in, so it covers no cell (FR-903); zero for none.
	Lane int
	// Beside is whether a pull out adjoins the ribbon, which then has a side for it (FR-901); PullOut
	// whether it is shown there rather than waiting behind its handle (FR-903).
	Beside, PullOut bool
}

// Host is the application whose ribbon is arranged.
type Host interface {
	// Ribbon answers the ribbon's choices and its content, read together so they cannot disagree.
	Ribbon() (ribbon.Choices, Content)
	// ChangeRibbon applies edit to the ribbon's choices through the application's own save path. A
	// save that fails keeps the change in effect and is answered; the application may raise a notice
	// for it, which can change its content (FR-707).
	ChangeRibbon(edit func(ribbon.Choices) ribbon.Choices) error
}

// Monitors answers the displays as the system reports them now.
type Monitors interface {
	Monitors() ([]placement.Monitor, error)
}

// Arrangement is where the window goes and how big it is, in physical pixels.
type Arrangement struct {
	At   placement.Point
	Size placement.Size
	// Scrolls is true when the cells need more room than the work area offers (FR-106).
	Scrolls bool
	// DPI is the DPI of the monitor the ribbon is on.
	DPI int
	// Edge is the edge running along the orientation the ribbon stands flush against; empty when it
	// stands against none, which leaves it pinned in effect (FR-619).
	Edge placement.Edge
	// PullOutSide is the ribbon's side the pull out and its handle go on; empty while there is no pull
	// out. PullOut is its rectangle; zero while none is shown (FR-902 to FR-904).
	PullOutSide placement.Edge
	PullOut     placement.Rect
}

// Arranger arranges one ribbon. It is safe to call from several goroutines. Its own lock is never
// held while it calls the host, which may hold a lock of its own while it calls in.
type Arranger struct {
	host       Host
	monitors   Monitors
	neighbours Neighbours

	mutex sync.Mutex
	// scrollbar is the thickness in DIP of the scroll bar the page draws, as the page measured it;
	// zero until it says (FR-106).
	scrollbar int
	// pixelsPerDIP is the scale the page is really drawn at, in window pixels to each DIP, as the
	// page reported it; zero until it says, when the display's DPI stands in for it.
	pixelsPerDIP float64
	// previewScale is the scale the ribbon is drawn at while its grip is dragged, kept nowhere and
	// not rounded; zero while no drag is under way (FR-623).
	previewScale float64
	// held is the ribbon and its pull out as they stood when a drag of the grip began, so the pull out
	// is held there while the drag lasts (FR-623).
	held heldPullOut
	// arranged is the ribbon's length when it was last arranged, so a change of length can be told
	// from anything else that arranges it (FR-104).
	arranged ribbonLength
	// last is where the ribbon was last arranged, so a ribbon placed again can keep the edge it lay
	// against (FR-408, FR-610).
	last lastPlaced
}

// lastPlaced is where the ribbon was last arranged: the display, its corner and its size, in that
// display's pixels; known is false until it has been arranged once.
type lastPlaced struct {
	known  bool
	device string
	at     placement.Point
	size   placement.Size
	// pullOut is the pull out's rectangle then; the zero rectangle while none was shown.
	pullOut placement.Rect
}

// heldPullOut is the ribbon and its pull out when a drag of the grip began; known is false while no
// drag is under way or no pull out was shown when it began.
type heldPullOut struct {
	known   bool
	ribbon  placement.Rect
	pullOut placement.Rect
}

// ribbonLength is the ribbon's length in DIP along its orientation; known is false until the ribbon
// has been arranged once, when there is nothing yet for a length to differ from. scale is the percent
// it was drawn at, so a length changed by the scale can be told from one changed by its content.
type ribbonLength struct {
	known    bool
	vertical bool
	length   int
	scale    float64
}

// New answers an arranger of host's ribbon on monitors, keeping it off the ribbons neighbours knows
// of (FR-412); nil neighbours is a ribbon alone.
func New(host Host, monitors Monitors, neighbours Neighbours) *Arranger {
	if neighbours == nil {
		neighbours = NoNeighbours{}
	}
	return &Arranger{host: host, monitors: monitors, neighbours: neighbours}
}

// SetScrollbar records the thickness in DIP of the scroll bar the page draws, which a scrolling
// ribbon makes room for across its cells (FR-106). Only the page can measure it: it is the web
// engine's bar, not one the system reports.
func (a *Arranger) SetScrollbar(dip int) error {
	if dip < 0 {
		return fmt.Errorf("%w: a scroll bar of %d", placement.ErrNegativeLength, dip)
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.scrollbar = dip
	return nil
}

// SetPixelsPerDIP records the scale the page is really drawn at, in window pixels to each DIP.
// Every size is converted with it from then on rather than with the display's DPI, which Windows'
// text size leaves as it was while it enlarges the page, so the window always fits the page.
func (a *Arranger) SetPixelsPerDIP(scale float64) error {
	if !(scale > 0) || math.IsInf(scale, 1) {
		return fmt.Errorf("%w: %v", ErrUnusableScale, scale)
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.pixelsPerDIP = scale
	return nil
}

// perDIP answers the pixels to each DIP a window on monitor is sized with.
func (a *Arranger) perDIP(monitor placement.Monitor) float64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return sizingScale(a.pixelsPerDIP, monitor)
}

// sizingScale answers reported, the page's own scale, once the page has reported one; else the
// scale of monitor's DPI.
func sizingScale(reported float64, monitor placement.Monitor) float64 {
	if reported > 0 {
		return reported
	}
	return placement.PerDIPOf(monitor.DPI)
}

// displays answers the monitors reported now; an error where none is or they cannot be read.
func (a *Arranger) displays() ([]placement.Monitor, error) {
	monitors, err := a.monitors.Monitors()
	if err != nil {
		return nil, fmt.Errorf("reading the displays: %w", err)
	}
	if len(monitors) == 0 {
		return nil, ErrNoMonitors
	}
	return monitors, nil
}

// Centred answers a window of size, in DIP, centred on the work area of the monitor holding at
// (CON-6): where a panel opens, since it shares the ribbon's window. Nothing is saved.
func (a *Arranger) Centred(at placement.Point, size placement.Size) (Arrangement, error) {
	monitors, err := a.displays()
	if err != nil {
		return Arrangement{}, err
	}
	monitor := mostOverlapped(monitors, at)
	work := monitor.Work
	// Never larger than the work area, so a short display still shows the whole surface.
	perDIP := a.perDIP(monitor)
	pixels := placement.Size{
		Width:  min(placement.PixelsOf(size.Width, perDIP), work.Width()),
		Height: min(placement.PixelsOf(size.Height, perDIP), work.Height()),
	}
	centre := placement.Point{X: work.Left + (work.Width()-pixels.Width)/2, Y: work.Top + (work.Height()-pixels.Height)/2}
	return Arrangement{At: placement.Clamp(centre, pixels, work), Size: pixels, DPI: monitor.DPI}, nil
}

// Collapsed answers the arrangement of the tab an unpinned ribbon arranged as full shrinks to
// (FR-614): the band of placement.Tab on the edge it stands flush against, its thickness in that
// display's pixels. Nothing is saved: the stored place stays the full ribbon's. A ribbon flush
// against no edge never collapses (FR-619), so asking for its tab is refused.
func (a *Arranger) Collapsed(full Arrangement) (Arrangement, error) {
	if full.Edge == "" {
		return Arrangement{}, ErrNotAgainstAnEdge
	}
	monitors, err := a.displays()
	if err != nil {
		return Arrangement{}, err
	}
	monitor := mostOverlapped(monitors, full.At)
	thickness := placement.PixelsOf(placement.TabThickness, a.perDIP(monitor))
	band := placement.Tab(full.At, full.Size, full.Edge, thickness)
	return Arrangement{
		At:   placement.Point{X: band.Left, Y: band.Top},
		Size: placement.Size{Width: band.Width(), Height: band.Height()},
		DPI:  monitor.DPI,
	}, nil
}
