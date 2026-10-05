package placement

import "testing"

// Two monitors side by side. The primary is 1920 by 1080 at 100 percent with a 48 pixel taskbar;
// the secondary is 2560 by 1440 at 150 percent, to its right.
var (
	primary = Monitor{
		Device: `\\.\DISPLAY1`, Work: Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1032},
		DPI: BaseDPI, Primary: true,
	}
	secondary = Monitor{
		Device: `\\.\DISPLAY2`, Work: Rect{Left: 1920, Top: 0, Right: 4480, Bottom: 1392},
		DPI: 144,
	}
	ribbon = Size{Width: 600, Height: 120}
)

// FR-403.
func TestDefaultPlacementIsRightEdgeCentred(t *testing.T) {
	t.Parallel()
	got := Default(primary, ribbon, Right)
	want := Point{X: 1920 - 600, Y: (1032 - 120) / 2}
	if got.At != want || got.Monitor.Device != primary.Device {
		t.Errorf("got %+v, want %+v on the primary", got.At, want)
	}
	scaled := Default(secondary, ribbon, Right)
	if gap := secondary.Work.Right - (scaled.At.X + ribbon.Width); gap != 0 {
		t.Errorf("at 150 percent the ribbon is %d pixels in from the right edge, want flush", gap)
	}
	// FR-409: the default place is against the home edge given, so a horizontal ribbon's is the top,
	// whether it has nothing stored, its monitor has gone or it was left off every display.
	top := Point{X: (1920 - 600) / 2, Y: 0}
	if got := Default(primary, ribbon, Top); got.At != top {
		t.Errorf("top: got %+v, want %+v", got.At, top)
	}
	gone := Stored{Device: `\\.\DISPLAY9`, DPI: BaseDPI}
	if got, _ := Restore(&gone, []Monitor{primary}, ribbon, Top); got.At != top {
		t.Errorf("a gone monitor with the top as home: got %+v, want %+v", got.At, top)
	}
	if got, _ := Recover(Point{X: 9000, Y: 0}, ribbon, []Monitor{primary}, Top); got.At != top {
		t.Errorf("off every display with the top as home: got %+v, want %+v", got.At, top)
	}
}

// FR-104: centred along its length on the work area, its position across kept, within the work
// area even when that position was off it.
func TestCentredAlongKeepsThePositionAcross(t *testing.T) {
	t.Parallel()
	tall := Size{Width: 120, Height: 600}
	if got, want := CentredAlong(Point{X: 1700, Y: 9}, tall, primary.Work, true), (Point{X: 1700, Y: (1032 - 600) / 2}); got != want {
		t.Errorf("vertical: got %+v, want %+v", got, want)
	}
	if got, want := CentredAlong(Point{X: 9, Y: 800}, ribbon, primary.Work, false), (Point{X: (1920 - 600) / 2, Y: 800}); got != want {
		t.Errorf("horizontal: got %+v, want %+v", got, want)
	}
	if got := CentredAlong(Point{X: 5000, Y: 0}, tall, primary.Work, true); got.X != 1920-120 {
		t.Errorf("a ribbon across the edge was not brought inside: %+v", got)
	}
}

// FR-408: flush against the edge, centred along it; on a work area that does not start at zero.
func TestAgainstEdgeIsFlushAndCentredAlongTheEdge(t *testing.T) {
	t.Parallel()
	tall := Size{Width: 120, Height: 600}
	work := secondary.Work
	cases := map[Edge]struct {
		size Size
		want Point
	}{
		Left:   {tall, Point{X: 1920, Y: (1392 - 600) / 2}},
		Right:  {tall, Point{X: 4480 - 120, Y: (1392 - 600) / 2}},
		Top:    {ribbon, Point{X: 1920 + (2560-600)/2, Y: 0}},
		Bottom: {ribbon, Point{X: 1920 + (2560-600)/2, Y: 1392 - 120}},
	}
	for edge, each := range cases {
		if got := AgainstEdge(each.size, work, edge); got != each.want {
			t.Errorf("%s: got %+v, want %+v", edge, got, each.want)
		}
	}
	long := Size{Width: 120, Height: 2000}
	if got := AgainstEdge(long, primary.Work, Right); got != (Point{X: 1920 - 120, Y: 0}) {
		t.Errorf("a ribbon longer than the work area was not aligned to its top: %+v", got)
	}
}

// FR-408, FR-610: a ribbon flush against the right or bottom edge keeps it as it shrinks or grows;
// one whose corner moved along that axis keeps its corner, as does one that was not flush.
func TestKeptFlushHoldsTheFarEdgeNotTheCorner(t *testing.T) {
	t.Parallel()
	work := primary.Work
	wide, narrow := Size{Width: 176, Height: 600}, Size{Width: 136, Height: 400}
	right := Point{X: 1920 - 176, Y: 200}
	if got := KeptFlush(right, narrow, right, wide, work); got != (Point{X: 1920 - 136, Y: 200}) {
		t.Errorf("shrinking on the right: got %+v", got)
	}
	if got := KeptFlush(Point{X: 1920 - 136, Y: 200}, wide, Point{X: 1920 - 136, Y: 200}, narrow, work); got != right {
		t.Errorf("growing on the right: got %+v, want %+v", got, right)
	}
	tall, short := Size{Width: 600, Height: 106}, Size{Width: 400, Height: 76}
	bottom := Point{X: 700, Y: 1032 - 106}
	if got := KeptFlush(bottom, short, bottom, tall, work); got != (Point{X: 700, Y: 1032 - 76}) {
		t.Errorf("shrinking on the bottom: got %+v", got)
	}
	moved := Point{X: 1500, Y: 200}
	if got := KeptFlush(moved, narrow, right, wide, work); got != moved {
		t.Errorf("a ribbon moved off the edge was pulled back: %+v", got)
	}
	free := Point{X: 900, Y: 200}
	if got := KeptFlush(free, narrow, free, wide, work); got != free {
		t.Errorf("a ribbon against no edge was moved: %+v", got)
	}
}

// FR-404, FR-405.
func TestPlacementIsStoredRelativeToItsMonitorAndRestored(t *testing.T) {
	t.Parallel()
	at := Point{X: 2100, Y: 300}
	stored := Record(at, secondary)
	if stored.Offset != (Point{X: 180, Y: 300}) || stored.Device != secondary.Device || stored.DPI != 144 {
		t.Fatalf("stored %+v", stored)
	}
	got, ok := Restore(&stored, []Monitor{primary, secondary}, ribbon, Right)
	if !ok || got.At != at || got.Monitor.Device != secondary.Device {
		t.Errorf("restored %+v on %s, want %+v on the secondary", got.At, got.Monitor.Device, at)
	}
}

// FR-405: the spec's acceptance example.
func TestMissingMonitorFallsBackToPrimary(t *testing.T) {
	t.Parallel()
	stored := Stored{Device: `\\.\DISPLAY2`, DPI: BaseDPI, Offset: Point{X: 1700, Y: 500}}
	got, ok := Restore(&stored, []Monitor{primary}, ribbon, Right)
	if !ok || got != Default(primary, ribbon, Right) {
		t.Errorf("got %+v, want the default place on the primary", got)
	}
}

func TestNothingStoredMeansTheDefaultPlace(t *testing.T) {
	t.Parallel()
	got, ok := Restore(nil, []Monitor{secondary, primary}, ribbon, Right)
	if !ok || got != Default(primary, ribbon, Right) {
		t.Errorf("got %+v, want the default place on the primary", got)
	}
}

// FR-405.
func TestOffscreenPlacementIsClampedIntoWorkArea(t *testing.T) {
	t.Parallel()
	stored := Stored{Device: primary.Device, DPI: BaseDPI, Offset: Point{X: 1800, Y: -40}}
	got, _ := Restore(&stored, []Monitor{primary}, ribbon, Right)
	if got.At != (Point{X: 1920 - 600, Y: 0}) {
		t.Errorf("got %+v", got.At)
	}
}

// FR-405.
func TestDpiChangeScalesTheOffset(t *testing.T) {
	t.Parallel()
	stored := Stored{Device: secondary.Device, DPI: BaseDPI, Offset: Point{X: 200, Y: 100}}
	got, _ := Restore(&stored, []Monitor{secondary}, ribbon, Right)
	if got.At != (Point{X: 1920 + 300, Y: 150}) {
		t.Errorf("got %+v, want the offset scaled by 144/96", got.At)
	}
}

// FR-406.
func TestDisplayChangeRecoversARibbonLeftOffscreen(t *testing.T) {
	t.Parallel()
	got, ok := Recover(Point{X: 3000, Y: 200}, ribbon, []Monitor{primary}, Right)
	if !ok || got != Default(primary, ribbon, Right) {
		t.Errorf("a ribbon on no monitor goes to the default place: got %+v", got)
	}
	half, _ := Recover(Point{X: 1700, Y: 200}, ribbon, []Monitor{primary, secondary}, Right)
	if half.Monitor.Device != secondary.Device || half.At.X != secondary.Work.Left {
		t.Errorf("a ribbon mostly on the secondary is clamped onto it: got %+v", half)
	}
}

func TestNoMonitorsIsReportedRatherThanGuessed(t *testing.T) {
	t.Parallel()
	if _, ok := Restore(nil, nil, ribbon, Right); ok {
		t.Error("Restore with no monitors answered a place")
	}
	if _, ok := Recover(Point{}, ribbon, nil, Right); ok {
		t.Error("Recover with no monitors answered a place")
	}
}

func TestPrimaryIsTheFirstWhenNoneIsMarked(t *testing.T) {
	t.Parallel()
	unmarked := secondary
	got, ok := Primary([]Monitor{unmarked})
	if !ok || got.Device != unmarked.Device {
		t.Errorf("got %+v", got)
	}
}

func TestARectangleMeasuresItsOwnSides(t *testing.T) {
	t.Parallel()
	if w, h := secondary.Work.Width(), secondary.Work.Height(); w != 2560 || h != 1392 {
		t.Errorf("got %d by %d", w, h)
	}
}

func TestARibbonLargerThanTheWorkAreaAlignsToItsStart(t *testing.T) {
	t.Parallel()
	got := Clamp(Point{X: 500, Y: 500}, Size{Width: 3000, Height: 2000}, primary.Work)
	if got != (Point{X: 0, Y: 0}) {
		t.Errorf("got %+v", got)
	}
}

func TestScaleRoundsToTheNearestPixelEitherSide(t *testing.T) {
	t.Parallel()
	cases := []struct{ length, from, to, want int }{
		{16, BaseDPI, 144, 24},
		{-200, BaseDPI, 144, -300},
		{5, BaseDPI, 120, 6},
		{-5, BaseDPI, 120, -6},
		{100, 0, 144, 150},
	}
	for _, each := range cases {
		if got := Scale(each.length, each.from, each.to); got != each.want {
			t.Errorf("Scale(%d, %d, %d) = %d, want %d", each.length, each.from, each.to, got, each.want)
		}
	}
}

// FR-105, FR-106, FR-107.
func TestRibbonLengthFollowsClockCountAndNeverExceedsWorkArea(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                       string
		cells, cell, padding, room int
		want                       Fitted
	}{
		{"three cells fit", 3, 200, 8, 1920, Fitted{Length: 616}},
		{"no cells keeps room for the prompt", 0, 200, 8, 1920, Fitted{Length: 216}},
		{"twelve cells overflow", 12, 200, 0, 1920, Fitted{Length: 1920, Scrolls: true}},
	}
	for _, each := range cases {
		if got := Fit(each.cells, each.cell, each.padding, each.room); got != each.want {
			t.Errorf("%s: got %+v, want %+v", each.name, got, each.want)
		}
	}
}
