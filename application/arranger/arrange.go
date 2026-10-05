package arranger

import (
	"math"
	"slices"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// Launch arranges the ribbon at launch or as a panel closes: on its stored monitor and offset, else
// at the default place on the primary monitor, always wholly inside a work area (FR-403, FR-405);
// kept against the right or bottom edge it lay against (FR-610); re-centred along its length where
// that length changed while the panel was open (FR-104).
func (a *Arranger) Launch() (Arrangement, error) {
	return a.recentredKept(func() (Arrangement, placement.Monitor, bool, error) {
		return a.arrange(func(monitors []placement.Monitor, current ribbon.Choices) placement.Monitor {
			return storedOrPrimary(monitors, current.Placement)
		}, func(size placement.Size, monitors []placement.Monitor, current ribbon.Choices) placement.Placed {
			placed, _ := placement.Restore(current.Placement, monitors, size, homeOf(current))
			return a.keptFlush(placed, size)
		})
	})
}

// Rearrange arranges a ribbon now at at after its content or the displays changed: the same
// top-left corner, moved the least distance that keeps it wholly inside the work area it overlaps
// most (FR-406); re-centred along its length on that work area where its content changed that length
// (FR-104), though not where its scale did (FR-623).
// A re-centring is saved, as is a move on the display its place was saved on; a move onto another
// display is not, so a monitor that comes back finds its placement kept.
func (a *Arranger) Rearrange(at placement.Point) (Arrangement, error) {
	return a.recentredKept(func() (Arrangement, placement.Monitor, bool, error) {
		return a.arrange(func(monitors []placement.Monitor, _ ribbon.Choices) placement.Monitor {
			return mostOverlapped(monitors, at)
		}, func(size placement.Size, monitors []placement.Monitor, current ribbon.Choices) placement.Placed {
			placed, _ := placement.Recover(at, size, monitors, homeOf(current))
			return a.keptFlush(placed, size)
		})
	})
}

// Moved records where a drag left the ribbon (FR-404) and answers where it belongs: back wholly on
// the work area it overlaps most where the drag left part of it off (FR-406), then flush against an
// edge along its orientation where its side came within placement.SnapReach of one (FR-410).
func (a *Arranger) Moved(at placement.Point) (Arrangement, error) {
	arranged, monitor, _, err := a.arrange(func(monitors []placement.Monitor, _ ribbon.Choices) placement.Monitor {
		return mostOverlapped(monitors, at)
	}, func(size placement.Size, monitors []placement.Monitor, current ribbon.Choices) placement.Placed {
		placed, _ := placement.Recover(at, size, monitors, homeOf(current))
		placed = a.keptFlush(placed, size)
		reach := placement.PixelsOf(placement.SnapReach, a.perDIP(placed.Monitor))
		placed.At = placement.Snapped(placed.At, size, placed.Monitor.Work, current.Orientation == ribbon.Vertical, reach)
		return placed
	})
	if err != nil {
		return arranged, err
	}
	return arranged, a.record(arranged.At, monitor)
}

// ToEdge puts a ribbon now at at flush against edge of the work area it overlaps most, centred along
// that edge; it keeps that place (FR-408).
func (a *Arranger) ToEdge(at placement.Point, edge placement.Edge) (Arrangement, error) {
	return a.toEdgeOf(func(monitors []placement.Monitor) placement.Monitor { return mostOverlapped(monitors, at) }, edge)
}

// ToLastEdge puts a ribbon now at at flush against the edge it last stood against, centred along it,
// as unticking Pin ribbon away from every edge does (FR-613): the orientation's home edge where none
// is remembered or the remembered one runs across the orientation; the same edge of the display
// holding at where the remembered display is not present. It keeps that place.
func (a *Arranger) ToLastEdge(at placement.Point) (Arrangement, error) {
	current, _ := a.read()
	edge, device := homeOf(current), ""
	if last := current.LastEdge; last != nil && placement.Along(last.Edge, current.Orientation == ribbon.Vertical) {
		edge, device = last.Edge, last.Device
	}
	return a.toEdgeOf(func(monitors []placement.Monitor) placement.Monitor {
		if index := slices.IndexFunc(monitors, func(m placement.Monitor) bool { return m.Device == device }); index >= 0 {
			return monitors[index]
		}
		return mostOverlapped(monitors, at)
	}, edge)
}

// toEdgeOf puts the ribbon flush against edge of the work area of the monitor pick chooses, centred
// along that edge; it keeps that place.
func (a *Arranger) toEdgeOf(pick func([]placement.Monitor) placement.Monitor, edge placement.Edge) (Arrangement, error) {
	return a.recentredKept(func() (Arrangement, placement.Monitor, bool, error) {
		arranged, monitor, _, err := a.arrange(func(monitors []placement.Monitor, _ ribbon.Choices) placement.Monitor {
			return pick(monitors)
		}, func(size placement.Size, monitors []placement.Monitor, _ ribbon.Choices) placement.Placed {
			monitor := pick(monitors)
			return placement.Placed{At: placement.AgainstEdge(size, monitor.Work, edge), Monitor: monitor}
		})
		return arranged, monitor, true, err
	})
}

// recentredKept answers what arrange answers, saving the ribbon's place where arrange says it was
// moved, re-centred or put against an edge; likewise where it now stands elsewhere on the display its
// place was saved on. The next launch finds it there (FR-104, FR-408). A save that fails may raise a
// notice the host draws as one more cell (FR-707), so the ribbon is arranged once more to fit it;
// that arrangement is not saved again.
func (a *Arranger) recentredKept(arrange func() (Arrangement, placement.Monitor, bool, error)) (Arrangement, error) {
	arranged, monitor, recentred, err := arrange()
	if err != nil || !(recentred || a.movedOnItsDisplay(arranged.At, monitor)) {
		return arranged, err
	}
	if a.record(arranged.At, monitor) == nil {
		return arranged, nil
	}
	arranged, _, _, err = arrange()
	return arranged, err
}

// movedOnItsDisplay answers whether a ribbon arranged at at on monitor stands somewhere other than
// the place saved for it on that same display, as one kept against its edge while it shrinks or grows
// across its breadth does (FR-408, FR-610). A ribbon moved onto another display is not, so a display
// that comes back finds its place kept (FR-405).
func (a *Arranger) movedOnItsDisplay(at placement.Point, monitor placement.Monitor) bool {
	current, _ := a.read()
	stored := current.Placement
	return stored != nil && stored.Device == monitor.Device && stored.Offset != placement.Record(at, monitor).Offset
}

// record stores at as the ribbon's place on monitor.
func (a *Arranger) record(at placement.Point, monitor placement.Monitor) error {
	stored := placement.Record(at, monitor)
	return a.host.ChangeRibbon(func(current ribbon.Choices) ribbon.Choices {
		current.Placement = &stored
		return current
	})
}

// arrange answers the arrangement on the monitor pick chooses, with place deciding the position
// once the size on that monitor is known. Where place lands the ribbon on another monitor, it is
// sized again in that monitor's pixels and placed again, so a ribbon moved onto a display at other
// scaling is drawn at that display's size (FR-407). Where the ribbon's length differs from the
// last arrangement's, it is re-centred along its length, which it answers (FR-104).
func (a *Arranger) arrange(
	pick func([]placement.Monitor, ribbon.Choices) placement.Monitor,
	place func(placement.Size, []placement.Monitor, ribbon.Choices) placement.Placed,
) (Arrangement, placement.Monitor, bool, error) {
	monitors, err := a.displays()
	if err != nil {
		return Arrangement{}, placement.Monitor{}, false, err
	}
	current, content := a.read()
	sizing := a.sizingOf(current)
	monitor := pick(monitors, current)
	size, scrolls, length := ribbonSize(current, content, sizing, monitor)
	placed := place(size, monitors, current)
	if placed.Monitor.Device != monitor.Device {
		size, scrolls, length = ribbonSize(current, content, sizing, placed.Monitor)
		placed = place(size, monitors, current)
	}
	vertical := current.Orientation == ribbon.Vertical
	recentred := a.lengthChanged(ribbonLength{known: true, vertical: vertical, length: length, scale: sizing.scale})
	if recentred {
		placed.At = placement.CentredAlong(placed.At, size, placed.Monitor.Work, vertical)
	}
	placed.At = a.clearOfOthers(current, content, placed, size)
	arranged := Arrangement{At: placed.At, Size: size, Scrolls: scrolls, DPI: placed.Monitor.DPI}
	if edge, flush := placement.FlushAgainst(arranged.At, size, placed.Monitor.Work, vertical); flush {
		arranged.Edge = edge
		a.rememberEdge(current, placement.Against{Device: placed.Monitor.Device, Edge: edge})
	}
	if side, rect, shown := a.pullOutBeside(current, content, arranged.At, size, placed.Monitor, arranged.Edge); side != "" {
		arranged.PullOutSide = side
		if shown {
			arranged.PullOut = rect
		}
		a.rememberSide(current, side)
	}
	a.remember(lastPlaced{known: true, device: placed.Monitor.Device, at: arranged.At, size: size, pullOut: arranged.PullOut})
	a.neighbours.Hold(footprintOf(arranged.At, size, arranged.PullOut, arranged.PullOut != placement.Rect{}))
	return arranged, placed.Monitor, recentred, nil
}

// rememberEdge keeps against as the edge the ribbon last stood flush against, however it got there
// (FR-411); saved only when it differs from current's. A save that fails is the host's to raise
// (FR-707), which the next arrangement fits, so it is not answered here.
func (a *Arranger) rememberEdge(current ribbon.Choices, against placement.Against) {
	if last := current.LastEdge; last != nil && *last == against {
		return
	}
	_ = a.host.ChangeRibbon(func(current ribbon.Choices) ribbon.Choices {
		current.LastEdge = &against
		return current
	})
}

// rememberSide keeps side as the side the pull out is on, so the next arrangement keeps it (FR-902,
// Amendment 35); saved only when it differs from current's. A failed save is the host's to raise
// (FR-707), as rememberEdge's is.
func (a *Arranger) rememberSide(current ribbon.Choices, side placement.Edge) {
	if current.PullOutSide == side {
		return
	}
	_ = a.host.ChangeRibbon(func(current ribbon.Choices) ribbon.Choices {
		current.PullOutSide = side
		return current
	})
}

// keptFlush answers placed kept against the right or bottom edge it lay against when the ribbon was
// last arranged on the same display, so a ribbon that shrinks or grows there stays against it
// (FR-408, FR-610).
func (a *Arranger) keptFlush(placed placement.Placed, size placement.Size) placement.Placed {
	a.mutex.Lock()
	last := a.last
	a.mutex.Unlock()
	if last.known && last.device == placed.Monitor.Device {
		placed.At = placement.KeptFlush(placed.At, size, last.at, last.size, placed.Monitor.Work)
	}
	return placed
}

// remember records now as where the ribbon was last arranged.
func (a *Arranger) remember(now lastPlaced) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.last = now
}

// lengthChanged records now as the ribbon's length, answering whether it differs from the length
// recorded last; never the first time, when there was none. A length changed with the scale does
// not count: the ribbon grows or shrinks from its corner then, as a window being resized does, so
// the grip stays under the pointer (FR-623).
func (a *Arranger) lengthChanged(now ribbonLength) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	changed := a.arranged.known && a.arranged != now && a.arranged.scale == now.scale
	a.arranged = now
	return changed
}

// sizing is the arranger's own share of what decides the ribbon's size, read together under its lock.
type sizing struct {
	scrollbar    int
	pixelsPerDIP float64
	// scale is the percent the ribbon is drawn at on top of its size (FR-623); fractional while the
	// grip is dragged.
	scale float64
}

func (a *Arranger) sizingOf(current ribbon.Choices) sizing {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return sizing{scrollbar: a.scrollbar, pixelsPerDIP: a.pixelsPerDIP, scale: a.scaleOf(current.Scale)}
}

// ribbonSize answers the ribbon's size on monitor in physical pixels (FR-105, FR-106): the cells
// fitted along the orientation within the work area; one cell plus padding across it, plus the
// scroll bar's thickness when the cells scroll, so the bar never covers them, plus the content's
// lane, so a handle there never covers them either (FR-903). All but the scroll bar are drawn at
// the chosen scale, so each DIP of them takes scale percent of the pixels it otherwise would; the
// bar is the web engine's own and keeps its thickness (FR-623). It answers the length along the
// orientation in DIP too, scaled, which a move between scalings keeps; a change of content
// re-centres by it, a change of scale does not.
func ribbonSize(current ribbon.Choices, content Content, sizing sizing, monitor placement.Monitor) (placement.Size, bool, int) {
	along, across := content.Cell.Width, content.Cell.Height
	room := monitor.Work.Width()
	if current.Orientation == ribbon.Vertical {
		along, across = content.Cell.Height, content.Cell.Width
		room = monitor.Work.Height()
	}
	perDIP := sizingScale(sizing.pixelsPerDIP, monitor)
	scaled := perDIP * sizing.scale / ribbon.WholeScale
	available := placement.DIPOf(room, scaled)
	fitted := placement.Fit(content.Cells, along, content.Padding, available)
	thickness := across + 2*content.Padding + content.Lane
	length := placement.PixelsOf(fitted.Length, scaled)
	breadth := placement.PixelsOf(thickness, scaled)
	if fitted.Scrolls {
		breadth += placement.PixelsOf(sizing.scrollbar, perDIP)
	}
	lengthDIP := int(math.Round(float64(fitted.Length) * sizing.scale / ribbon.WholeScale))
	if current.Orientation == ribbon.Vertical {
		return placement.Size{Width: breadth, Height: length}, fitted.Scrolls, lengthDIP
	}
	return placement.Size{Width: length, Height: breadth}, fitted.Scrolls, lengthDIP
}

// pullOutBeside answers where the pull out goes for a ribbon arranged at at of size on monitor, flush
// against edge (none when empty): the side it adjoins (where its handle goes too), kept from the last
// arrangement while no edge decides it (FR-902, Amendment 35), with its rectangle;
// false for the rectangle while none is shown, the pull out closed included (FR-902 to FR-904). The
// side is empty while content has no pull out beside it.
func (a *Arranger) pullOutBeside(current ribbon.Choices, content Content, at placement.Point, size placement.Size, monitor placement.Monitor, edge placement.Edge) (placement.Edge, placement.Rect, bool) {
	if !content.Beside {
		return "", placement.Rect{}, false
	}
	vertical := current.Orientation == ribbon.Vertical
	ribbon := placement.Rect{Left: at.X, Top: at.Y, Right: at.X + size.Width, Bottom: at.Y + size.Height}
	perDIP := a.perDIP(monitor)
	floor := placement.PixelsOf(placement.PullOutFloor, perDIP)
	side := placement.PullOutSideOf(ribbon, monitor.Work, vertical, edge, current.PullOutSide, floor)
	if !content.PullOut {
		return side, placement.Rect{}, false
	}
	minimum := placement.PixelsOf(placement.PullOutMinimumWidth, perDIP)
	rect, shown := placement.PullOutBeside(ribbon, monitor.Work, side, minimum, floor)
	if held := a.heldOf(); held.known && shown {
		rect = placement.PullOutHeld(ribbon, side, held.pullOut, held.ribbon)
	}
	return side, rect, shown
}

// storedOrPrimary answers the stored monitor where it is present; else the primary.
func storedOrPrimary(monitors []placement.Monitor, stored *placement.Stored) placement.Monitor {
	if stored != nil {
		index := slices.IndexFunc(monitors, func(m placement.Monitor) bool { return m.Device == stored.Device })
		if index >= 0 {
			return monitors[index]
		}
	}
	primary, _ := placement.Primary(monitors)
	return primary
}

// mostOverlapped answers the monitor holding at; the primary when none does. Only the monitor is read,
// so the edge a ribbon on none would go to makes no difference.
func mostOverlapped(monitors []placement.Monitor, at placement.Point) placement.Monitor {
	placed, _ := placement.Recover(at, placement.Size{Width: 1, Height: 1}, monitors, placement.Right)
	return placed.Monitor
}

// homeOf answers the home edge of current's orientation (FR-409), where a ribbon with no place of its
// own goes. read normalises the choices, so the orientation is always one offered.
func homeOf(current ribbon.Choices) placement.Edge {
	edge, _ := ribbon.HomeEdge(current.Orientation)
	return edge
}
