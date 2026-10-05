package window

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
)

// Every change the page can make is followed by fitting the ribbon, whether or not it saved: a
// change whose save failed raises a notice, which is one more cell to fit (FR-707).
func TestEveryChangeFitsTheRibbonAndAnswersTheServicesError(t *testing.T) {
	changes := map[string]func(app *Window) error{
		"SetColour":       func(app *Window) error { return app.SetColour("neon") },
		"SetTheme":        func(app *Window) error { return app.SetTheme("dark") },
		"SetAlwaysOnTop":  func(app *Window) error { return app.SetAlwaysOnTop(true) },
		"SetScrollbar":    func(app *Window) error { return app.SetScrollbar(12) },
		"PreviewScale":    func(app *Window) error { return app.PreviewScale(150) },
		"SetScale":        func(app *Window) error { return app.SetScale(150) },
		"SetPixelsPerDIP": func(app *Window) error { return app.SetPixelRatio(1.25) },
	}
	for name, change := range changes {
		for _, failure := range []error{nil, errPlanted} {
			app, service, seen, _ := newTestApp(t)
			service.changeErr = failure
			if err := change(app); !errors.Is(err, failure) {
				t.Errorf("%s answered %v, want the service's %v", name, err, failure)
			}
			if !slices.Contains(service.calls, name) {
				t.Errorf("%s never reached the service", name)
			}
			if !slices.Contains(service.calls, "Rearrange") || len(seen.placed) != 1 {
				t.Errorf("%s (service answered %v) placed the ribbon %d times, want it fitted once", name, failure, len(seen.placed))
			}
		}
	}
}

// While a panel is open the window is that panel, so a change does not fit the ribbon; nor does one
// made before startup has found it.
func TestAChangeLeavesThePanelOrAnUnfoundRibbonAlone(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.panelOpen.Store(true)
	_ = app.SetScrollbar(12)
	app.panelOpen.Store(false)
	app.ribbon = 0
	_ = app.SetScrollbar(12)
	if slices.Contains(service.calls, "Rearrange") || len(seen.placed) != 0 {
		t.Errorf("the ribbon was fitted %d times, want none", len(seen.placed))
	}
}

func TestFittingPlacesTheArrangementAndKeepsWhetherItScrolls(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	_ = app.SetScrollbar(12)
	if service.at[0] != testRibbonAt {
		t.Errorf("fitted from %v, want where the ribbon stands, %v", service.at[0], testRibbonAt)
	}
	if seen.placed[0].At != testArrange.At || seen.placed[0].Size != testArrange.Size {
		t.Errorf("placed %+v, want the service's arrangement %+v", seen.placed[0], testArrange)
	}
	if !app.scrolls.Load() || !app.shown().Scrolls {
		t.Error("the arrangement scrolls but the window does not say so")
	}
}

func TestAFitThatCannotBeWorkedOutIsLoggedAndNothingMoves(t *testing.T) {
	for _, plant := range []func(*scriptedService, *window){
		func(_ *scriptedService, seen *window) { seen.readErr = errPlanted },
		func(service *scriptedService, _ *window) { service.arrangeErr = errPlanted },
	} {
		app, service, seen, log := newTestApp(t)
		plant(service, seen)
		_ = app.SetScrollbar(12)
		if len(seen.placed) != 0 || log.Len() == 0 {
			t.Errorf("placed %d times with log %q, want nothing placed and the failure logged", len(seen.placed), log)
		}
	}
}

func TestSetAlwaysOnTopAppliesTheSettingAtOnce(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	_ = app.SetAlwaysOnTop(true)
	if !slices.Equal(seen.onTop, []bool{true}) {
		t.Errorf("the window was set on top %v, want [true]", seen.onTop)
	}
}

func TestOpenPanelCentresThePanelOnTheRibbonsDisplay(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	if err := app.OpenPanel(openAtAbout); err != nil {
		t.Fatal(err)
	}
	if !app.panelOpen.Load() || service.at[0] != testRibbonAt || service.centred != testPanel {
		t.Errorf("open %v, centred %v from %v; want open, the panel's size from the ribbon", app.panelOpen.Load(), service.centred, service.at)
	}
	if len(seen.placed) != 1 || seen.placed[0].At != testArrange.At {
		t.Errorf("placed %+v, want the centred arrangement", seen.placed)
	}
}

func TestOpenPanelAnswersWhatStoppedIt(t *testing.T) {
	for _, plant := range []func(*scriptedService, *window){
		func(_ *scriptedService, seen *window) { seen.readErr = errPlanted },
		func(service *scriptedService, _ *window) { service.arrangeErr = errPlanted },
		func(_ *scriptedService, seen *window) { seen.placeErr = errPlanted },
	} {
		app, service, seen, _ := newTestApp(t)
		plant(service, seen)
		if err := app.OpenPanel(openAtAbout); !errors.Is(err, errPlanted) {
			t.Errorf("OpenPanel answered %v, want the failure", err)
		}
		if !app.panelOpen.Load() {
			t.Error("a panel that failed to place is no longer counted open, so the ribbon would be fitted over it")
		}
	}
}

func TestClosePanelPutsTheRibbonWhereItWasLastLeft(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.panelOpen.Store(true)
	if err := app.ClosePanel(); err != nil {
		t.Fatal(err)
	}
	if app.panelOpen.Load() || !slices.Contains(service.calls, "Launch") || len(seen.placed) != 1 {
		t.Errorf("open %v after %v, placed %d; want closed, launched and placed", app.panelOpen.Load(), service.calls, len(seen.placed))
	}
	service.arrangeErr = errPlanted
	if err := app.ClosePanel(); !errors.Is(err, errPlanted) {
		t.Errorf("ClosePanel answered %v, want the service's failure", err)
	}
}

func TestShowContextMenuShowsTheServicesMenu(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	service.menu = []menus.Item{{Action: menus.Hide, Label: "Hide ribbon"}}
	app.ShowContextMenu()
	if len(seen.menus) != 1 || !reflect.DeepEqual(seen.menus[0], service.menu) {
		t.Errorf("showed %v, want %v", seen.menus, service.menu)
	}
}

func TestOpenDonationHandsTheAddressToTheBrowser(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if err := app.OpenDonation(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(seen.browsed, []string{testProduct.DonateURL}) {
		t.Errorf("browsed %v, want the donation address once", seen.browsed)
	}
}

// A browser the desktop could not open is reported with the reason and the address, so the page can
// still be reached by hand rather than the button doing nothing. The words name no system.
func TestADonationPageThatCouldNotBeOpenedIsReportedWithItsAddress(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	seen.browseErr = errPlanted
	err := app.OpenDonation()
	if !errors.Is(err, errPlanted) {
		t.Fatalf("OpenDonation answered %v, want the desktop's refusal", err)
	}
	if !strings.Contains(err.Error(), testProduct.DonateURL) {
		t.Errorf("the refusal %q does not give the address", err)
	}
	for _, system := range []string{"Windows", "macOS", "Linux"} {
		if strings.Contains(err.Error(), system) {
			t.Errorf("the refusal %q names %s, though every platform can give it", err, system)
		}
	}
}

// What Help shows is the product the window was handed (FR-607, FR-608).
func TestHelpShowsTheProduct(t *testing.T) {
	app, _, _, _ := newTestApp(t)
	about := app.About()
	if about.Name != testProduct.App.Name || about.Version != testProduct.Version || about.Author != testProduct.Author || about.Copyright != testProduct.Copyright {
		t.Errorf("About answered %+v, want the product's", about)
	}
	if len(about.Credits) != 1 || about.Credits[0] != (creditDTO{Name: "Wails", Licence: "MIT", Role: "window"}) {
		t.Errorf("credits %+v, want the product's", about.Credits)
	}
	if app.Licence() != testProduct.Licence {
		t.Errorf("Licence answered %q, want the product's terms", app.Licence())
	}
}

func TestTheReadingsPassThrough(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	if on, err := app.StartWithWindows(); !on || err != nil {
		t.Errorf("StartWithWindows answered %v, %v", on, err)
	}
	if err := app.SetStartWithWindows(true); err != nil || !slices.Contains(service.calls, "SetStartWithWindows") {
		t.Errorf("SetStartWithWindows answered %v after %v", err, service.calls)
	}
	app.Hide()
	if app.visible.Load() {
		t.Error("Hide left the ribbon counted visible")
	}
}
