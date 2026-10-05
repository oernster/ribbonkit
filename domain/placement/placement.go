// Package placement decides where the ribbon goes: its default position, the position restored
// from what was stored and the recovery that brings it back wholly onto a work area.
//
// Every coordinate is in physical pixels on the virtual desktop, as Windows reports it. A length
// given in DIP is converted with the monitor's DPI.
package placement

import "slices"

// BaseDPI is the DPI of 100 percent scaling, where one DIP is one physical pixel. Windows names it
// USER_DEFAULT_SCREEN_DPI.
const BaseDPI = 96

// Rect is a rectangle whose right and bottom edges lie just outside it, as Windows RECT does.
type Rect struct {
	Left, Top, Right, Bottom int
}

// Width answers the rectangle's width.
func (r Rect) Width() int { return r.Right - r.Left }

// Height answers the rectangle's height.
func (r Rect) Height() int { return r.Bottom - r.Top }

// Point is a position on the virtual desktop.
type Point struct {
	X, Y int
}

// Size is a width and a height.
type Size struct {
	Width, Height int
}

// Monitor is one display as Windows reports it.
type Monitor struct {
	// Device is the monitor's device name, such as `\\.\DISPLAY2`.
	Device string
	// Work is the monitor's work area.
	Work Rect
	// DPI is the monitor's effective DPI.
	DPI int
	// Primary is true for the primary monitor.
	Primary bool
}

// Stored is the placement kept between runs (FR-404).
type Stored struct {
	Device string
	Work   Rect
	DPI    int
	// Offset is the ribbon's top-left corner relative to Work's top-left corner.
	Offset Point
}

// Placed is a decided position with the monitor it is on.
type Placed struct {
	At      Point
	Monitor Monitor
}

// Record answers what to store for a ribbon at at on monitor.
func Record(at Point, monitor Monitor) Stored {
	return Stored{
		Device: monitor.Device,
		Work:   monitor.Work,
		DPI:    monitor.DPI,
		Offset: Point{X: at.X - monitor.Work.Left, Y: at.Y - monitor.Work.Top},
	}
}

// Primary answers the primary monitor; the first when none says it is primary. It answers false
// when there are no monitors at all.
func Primary(monitors []Monitor) (Monitor, bool) {
	if len(monitors) == 0 {
		return Monitor{}, false
	}
	index := slices.IndexFunc(monitors, func(m Monitor) bool { return m.Primary })
	return monitors[max(index, 0)], true
}

// Default answers the default place on monitor for a ribbon of size (FR-403): flush against home,
// the edge of the ribbon's orientation (FR-409), centred along it.
func Default(monitor Monitor, size Size, home Edge) Placed {
	return Placed{At: AgainstEdge(size, monitor.Work, home), Monitor: monitor}
}

// CentredAlong answers at with a ribbon of size centred on work along its length (top to bottom
// when vertical, else left to right), its position across kept; then clamped (FR-104).
func CentredAlong(at Point, size Size, work Rect, vertical bool) Point {
	if vertical {
		at.Y = work.Top + (work.Height()-size.Height)/2
	} else {
		at.X = work.Left + (work.Width()-size.Width)/2
	}
	return Clamp(at, size, work)
}

// Edge is one side of a work area a ribbon can be put against (FR-408).
type Edge string

// The edges.
const (
	Left   Edge = "left"
	Right  Edge = "right"
	Top    Edge = "top"
	Bottom Edge = "bottom"
)

// AgainstEdge answers a ribbon of size flush against edge of work, centred along that edge: top to
// bottom for the left and right edges, left to right for the top and bottom (FR-408); then clamped,
// so a ribbon longer than work is aligned to its top or left.
func AgainstEdge(size Size, work Rect, edge Edge) Point {
	centred := Point{
		X: work.Left + (work.Width()-size.Width)/2,
		Y: work.Top + (work.Height()-size.Height)/2,
	}
	switch edge {
	case Left:
		centred.X = work.Left
	case Right:
		centred.X = work.Right - size.Width
	case Top:
		centred.Y = work.Top
	case Bottom:
		centred.Y = work.Bottom - size.Height
	}
	return Clamp(centred, size, work)
}

// KeptFlush answers at for a ribbon of size placed again on work, where it was last at was with size
// wasSize: a ribbon that lay flush against work's right or bottom edge, its corner not since moved
// along that axis, is kept flush against it, so a ribbon that shrinks or grows there keeps its far edge
// rather than its corner (FR-408, FR-610); then clamped. The left and top edges hold the corner, so
// they keep a ribbon without help.
func KeptFlush(at Point, size Size, was Point, wasSize Size, work Rect) Point {
	if at.X == was.X && was.X+wasSize.Width == work.Right {
		at.X = work.Right - size.Width
	}
	if at.Y == was.Y && was.Y+wasSize.Height == work.Bottom {
		at.Y = work.Bottom - size.Height
	}
	return Clamp(at, size, work)
}

// Restore answers where the ribbon goes at launch (FR-405): on the stored monitor at the stored
// offset scaled by the change in its DPI; else at the default place on the primary monitor. Either
// way it is clamped wholly inside the work area. home is the edge the default place is against. It
// answers false only when there are no monitors.
func Restore(stored *Stored, monitors []Monitor, size Size, home Edge) (Placed, bool) {
	primary, ok := Primary(monitors)
	if !ok {
		return Placed{}, false
	}
	if stored == nil {
		return Default(primary, size, home), true
	}
	index := slices.IndexFunc(monitors, func(m Monitor) bool { return m.Device == stored.Device })
	if index < 0 {
		return Default(primary, size, home), true
	}
	monitor := monitors[index]
	at := Point{
		X: monitor.Work.Left + Scale(stored.Offset.X, stored.DPI, monitor.DPI),
		Y: monitor.Work.Top + Scale(stored.Offset.Y, stored.DPI, monitor.DPI),
	}
	return Placed{At: Clamp(at, size, monitor.Work), Monitor: monitor}, true
}

// Recover answers where a ribbon now at at belongs after the displays changed (FR-406): clamped
// into the monitor it overlaps most; the default place on the primary monitor when it overlaps
// none, against home. It answers false only when there are no monitors.
func Recover(at Point, size Size, monitors []Monitor, home Edge) (Placed, bool) {
	primary, ok := Primary(monitors)
	if !ok {
		return Placed{}, false
	}
	ribbon := Rect{Left: at.X, Top: at.Y, Right: at.X + size.Width, Bottom: at.Y + size.Height}
	best, bestArea := -1, 0
	for index, monitor := range monitors {
		if area := overlap(ribbon, monitor.Work); area > bestArea {
			best, bestArea = index, area
		}
	}
	if best < 0 {
		return Default(primary, size, home), true
	}
	return Placed{At: Clamp(at, size, monitors[best].Work), Monitor: monitors[best]}, true
}

// Clamp answers at moved the least distance that brings a ribbon of size wholly inside work. A ribbon
// larger than work along an axis is aligned to work's left or top edge on that axis.
func Clamp(at Point, size Size, work Rect) Point {
	return Point{
		X: clampAxis(at.X, size.Width, work.Left, work.Right),
		Y: clampAxis(at.Y, size.Height, work.Top, work.Bottom),
	}
}

func clampAxis(position, length, low, high int) int {
	if position+length > high {
		position = high - length
	}
	return max(position, low)
}

func overlap(a, b Rect) int {
	width := min(a.Right, b.Right) - max(a.Left, b.Left)
	height := min(a.Bottom, b.Bottom) - max(a.Top, b.Top)
	if width <= 0 || height <= 0 {
		return 0
	}
	return width * height
}

// Scale converts length measured at DPI from to DPI to, rounding to the nearest pixel.
func Scale(length, from, to int) int {
	if from <= 0 {
		from = BaseDPI
	}
	product := length * to
	if product < 0 {
		return -((-product + from/2) / from)
	}
	return (product + from/2) / from
}
