package window

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// fittedHeight is a content height taller than the panel opens at.
const fittedHeight = 1200

// FR-621: an open panel is centred again at its content's height and the panel's own width, from
// where the window stands.
func TestFitPanelMakesTheOpenPanelAsTallAsItsContent(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	if err := app.OpenPanel(openAtAbout); err != nil {
		t.Fatal(err)
	}
	if err := app.FitPanel(fittedHeight); err != nil {
		t.Fatal(err)
	}
	if want := (placement.Size{Width: testPanel.Width, Height: fittedHeight}); service.centred != want || service.at[len(service.at)-1] != testRibbonAt {
		t.Errorf("centred %v from %v, want %v from the window", service.centred, service.at, want)
	}
	if len(seen.placed) != 2 || seen.placed[1].At != testArrange.At || seen.placed[1].Size != testArrange.Size {
		t.Errorf("placed %+v, want the panel then its fitted arrangement", seen.placed)
	}
}

// FR-621: with no panel open nothing is fitted, nor with no height; a negative height is refused.
func TestFitPanelLeavesTheRibbonAloneAndRefusesANegativeHeight(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if err := app.FitPanel(fittedHeight); err != nil || len(seen.placed) != 0 {
		t.Errorf("no panel: %v, placed %+v", err, seen.placed)
	}
	app.panelOpen.Store(true)
	if err := app.FitPanel(0); err != nil || len(seen.placed) != 0 {
		t.Errorf("no height: %v, placed %+v", err, seen.placed)
	}
	if err := app.FitPanel(-1); !errors.Is(err, placement.ErrNegativeLength) {
		t.Errorf("a negative height answered %v", err)
	}
}

// FR-621: what stopped the fit is answered.
func TestFitPanelAnswersWhatStoppedIt(t *testing.T) {
	for _, plant := range []func(*scriptedService, *window){
		func(_ *scriptedService, seen *window) { seen.readErr = errPlanted },
		func(service *scriptedService, _ *window) { service.arrangeErr = errPlanted },
		func(_ *scriptedService, seen *window) { seen.placeErr = errPlanted },
	} {
		app, service, seen, _ := newTestApp(t)
		app.panelOpen.Store(true)
		plant(service, seen)
		if err := app.FitPanel(fittedHeight); !errors.Is(err, errPlanted) {
			t.Errorf("FitPanel answered %v, want the failure", err)
		}
	}
}
