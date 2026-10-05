package window

import (
	"context"
	"strings"
	"testing"
)

// The page's sizing may land before it is ready: the ribbon is then shown the moment it is, with no
// fallback waited for.
func TestARibbonSizedBeforeItIsReadyShowsWhenReady(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if err := app.SetPixelRatio(1); err != nil {
		t.Fatal(err)
	}
	app.pageMeasured()
	if seen.shown != 0 {
		t.Fatalf("shown %d times before the page was ready", seen.shown)
	}
	app.domReady(context.Background())
	if seen.shown != 1 || seen.sizePending != nil {
		t.Errorf("shown %d times with a fallback pending %v, want shown once at once", seen.shown, seen.sizePending != nil)
	}
}

// A page that never sizes the ribbon still shows it once the fallback falls due; a report that comes
// late does not show it again.
func TestARibbonThePageNeverSizesIsShownByTheFallbackOnce(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.domReady(context.Background())
	if seen.sizePending == nil {
		t.Fatal("no fallback waited for")
	}
	seen.sizePending()
	if seen.shown != 1 {
		t.Fatalf("shown %d times by the fallback, want once", seen.shown)
	}
	if err := app.SetPixelRatio(1); err != nil {
		t.Fatal(err)
	}
	app.pageMeasured()
	if seen.shown != 1 {
		t.Errorf("a late report showed the ribbon again: %d times", seen.shown)
	}
}

// showLaunchedRibbon places the ribbon as a launch does, then has the page make it ready and size
// it, which shows it.
func showLaunchedRibbon(t *testing.T, app *Window) {
	t.Helper()
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.domReady(context.Background())
	if err := app.SetPixelRatio(1); err != nil {
		t.Fatal(err)
	}
	app.pageMeasured()
}

// FR-403, FR-405: a desktop that shows the launched window somewhere else (GNOME, measured
// 2026-10-04) has it put back where it was placed, once, said in the log; one that shows it where it
// was placed is left alone. The page's reports refit the window before it is shown, so the two are
// told apart by the one placement the first makes more.
func TestALaunchedRibbonShownElsewhereIsPlacedAgain(t *testing.T) {
	inPlace, _, seenInPlace, logInPlace := newTestApp(t)
	at, _, _ := windowOf(testArrange)
	seenInPlace.ribbonAt = at
	showLaunchedRibbon(t, inPlace)
	if strings.Contains(logInPlace.String(), "placing it again") {
		t.Errorf("a window shown where it was placed was placed again:\n%s", logInPlace)
	}

	elsewhere, _, seen, log := newTestApp(t)
	showLaunchedRibbon(t, elsewhere)
	if len(seen.placed) != len(seenInPlace.placed)+1 {
		t.Fatalf("placed %d times, want one more than the %d of a window shown in place", len(seen.placed), len(seenInPlace.placed))
	}
	if last := seen.placed[len(seen.placed)-1].At; last != at {
		t.Errorf("placed again at %+v, want where the launch put it, %+v", last, at)
	}
	if !strings.Contains(log.String(), "placing it again") {
		t.Errorf("the log does not say the window was placed again:\n%s", log)
	}
}

// A scale the service refuses does not count as sizing the ribbon, measured or not; a refused
// measurement never reaches the window as measured (the application's own test).
func TestARefusedReportDoesNotShowTheRibbon(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.domReady(context.Background())
	service.changeErr = errPlanted
	_ = app.SetPixelRatio(1)
	app.pageMeasured()
	if seen.shown != 0 {
		t.Errorf("refused reports showed the ribbon %d times", seen.shown)
	}
}
