package window

import (
	"math"
	"slices"
	"testing"
)

// The page's ratio is turned into window pixels by the desktop with the toolkit's own window scale it
// reports, so a desktop that scales windows itself is not scaled twice; the desktop's answer is what
// windows are sized with.
func TestThePagesRatioIsTakenWithTheToolkitsScale(t *testing.T) {
	t.Parallel()
	app, _, seen, _ := newTestApp(t)
	const ratio, toolkit, answered = 3, 2, 1.5
	seen.toolkitScale, seen.perDIP = toolkit, answered
	if err := app.SetPixelRatio(ratio); err != nil {
		t.Fatal(err)
	}
	if want := []perDIPAsk{{ratio, toolkit}}; !slices.Equal(seen.perDIPAsked, want) {
		t.Errorf("the desktop was asked %v, want %v", seen.perDIPAsked, want)
	}
	if got := math.Float64frombits(app.pixelsPerDIP.Load()); got != answered {
		t.Errorf("windows are sized at %v pixels to a unit, want the desktop's %v", got, answered)
	}
}
