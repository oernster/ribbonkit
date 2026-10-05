package window

import (
	"math"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// FR-623: the grip's page cannot take a snapshot mid-drag by itself, so a preview and the kept
// scale each tell it to draw again, as well as fitting the window (the fit is proved in
// TestEveryChangeFitsTheRibbonAndAnswersTheServicesError).
func TestAChangeOfScaleTellsThePageToDrawAgain(t *testing.T) {
	for name, change := range map[string]func(*Window) error{
		"PreviewScale": func(app *Window) error { return app.PreviewScale(150) },
		"SetScale":     func(app *Window) error { return app.SetScale(150) },
	} {
		app, _, seen, _ := newTestApp(t)
		if err := change(app); err != nil || !seen.sawEvent(eventRefresh) {
			t.Errorf("%s answered %v; refresh sent %v", name, err, seen.sawEvent(eventRefresh))
		}
	}
}

// gripThickness is the ribbon's thickness in the page's units when a test's drag begins.
const gripThickness = 200.0

// FR-623: where the desktop can read the pointer, the drag follows it in the page's units;
// never the page's own reading, which was measured jumping backwards mid-drag (2026-10-04).
func TestTheGripFollowsTheDesktopsPointerOverThePages(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.choices.Scale, service.choices.Orientation = ribbon.WholeScale, ribbon.Vertical
	const perDIP = 2.0
	app.pixelsPerDIP.Store(math.Float64bits(perDIP))
	cursor := placement.Point{X: 1000, Y: 300}
	app.cursor = func() (placement.Point, bool) { return cursor, true }
	app.BeginScale(gripThickness, 0, 0)
	cursor.X += int(gripThickness / 4 * perDIP)
	if err := app.DragScale(-gripThickness, 0); err != nil {
		t.Fatal(err)
	}
	if err := app.EndScale(-gripThickness, 0); err != nil {
		t.Fatal(err)
	}
	want := ribbon.WholeScale + ribbon.WholeScale/4
	if !slices.Equal(service.previewed, []float64{float64(want)}) || !slices.Equal(service.kept, []int{want}) {
		t.Errorf("previewed %v, kept %v; want %d for both", service.previewed, service.kept, want)
	}
}

// FR-623: where the desktop cannot read the pointer, the page's reading moves the far side: down
// for a horizontal ribbon; the same scale twice is previewed once.
func TestTheGripFollowsThePagesPointerWhereTheDesktopCannotReadIt(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.choices.Scale, service.choices.Orientation = ribbon.WholeScale, ribbon.Horizontal
	app.BeginScale(gripThickness, 0, 0)
	for range 2 {
		if err := app.DragScale(gripThickness, gripThickness/2); err != nil {
			t.Fatal(err)
		}
	}
	want := ribbon.WholeScale + ribbon.WholeScale/2
	if !slices.Equal(service.previewed, []float64{float64(want)}) {
		t.Errorf("previewed %v, want only %d", service.previewed, want)
	}
}

// FR-623: a press that previews nothing keeps nothing, as a double-click's presses do; a call
// after the drag has ended does nothing.
func TestAPressThatMovesNothingKeepsNothing(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.choices.Scale = ribbon.WholeScale
	app.BeginScale(gripThickness, 0, 0)
	if err := app.DragScale(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := app.EndScale(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := app.DragScale(0, gripThickness); err != nil {
		t.Fatal(err)
	}
	if err := app.EndScale(0, gripThickness); err != nil {
		t.Fatal(err)
	}
	if len(service.previewed) != 0 || len(service.kept) != 0 {
		t.Errorf("previewed %v, kept %v; want nothing", service.previewed, service.kept)
	}
}
