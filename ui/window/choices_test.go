package window

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// testChoices is a Colour group of two schemes, Neon ticked, then Pin ribbon on its own.
var testChoices = []menus.Item{
	{Label: "Colour", Children: []menus.Item{
		{Action: "colour-classic", Label: "Classic", Checkable: true},
		{Action: "colour-neon", Label: "Neon", Checkable: true, Checked: true},
	}},
	{Action: menus.Pin, Label: "Pin ribbon", Checkable: true, Checked: true},
}

// FR-624: a choice made in Settings is carried out as its menu item is: one of the kit's by the
// window, one of the application's handed to it. Anything else is refused and changes nothing, a
// group's own label included, since it chooses nothing.
func TestChooseCarriesOutOnlyTheChoicesSettingsOffers(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	service.settingsFor = testChoices
	if err := app.Choose(string(menus.Pin)); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(service.calls, "SetPinned") || !seen.sawEvent(eventRefresh) {
		t.Errorf("calls %v; want the pin changed and the page told", service.calls)
	}
	if err := app.Choose("colour-classic"); err != nil || !slices.Equal(seen.acted, []menus.Action{"colour-classic"}) {
		t.Errorf("Choose answered %v, handing over %v; want the application's choice handed to it", err, seen.acted)
	}
	for _, refused := range []string{"", "exit", "colour-sunset"} {
		app, service, seen, _ := newTestApp(t)
		service.settingsFor = testChoices
		if err := app.Choose(refused); !errors.Is(err, ribbon.ErrUnknownChoice) || len(service.calls) != 0 || len(seen.acted) != 0 {
			t.Errorf("Choose(%q) answered %v with calls %v; want it refused and nothing done", refused, err, service.calls)
		}
	}
}

// FR-625: Settings opens at its own width and keeps it as its height is fitted; the other panels at
// theirs. A word naming no panel is refused and opens nothing.
func TestSettingsOpensAndFitsAtItsOwnWidth(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	if err := app.OpenPanel(openAtSettings); err != nil || service.centred != testSettingsPanel {
		t.Fatalf("Settings centred at %v (%v), want %v", service.centred, err, testSettingsPanel)
	}
	if err := app.FitPanel(fittedHeight); err != nil || service.centred.Width != testSettingsPanel.Width {
		t.Errorf("fitted at %v (%v), want Settings' width kept", service.centred, err)
	}
	for _, other := range []string{openAtAbout, openAtLicence, openAtUpdate} {
		app, service, _, _ := newTestApp(t)
		if err := app.OpenPanel(other); err != nil || service.centred != testPanel {
			t.Errorf("%s centred at %v (%v), want %v", other, service.centred, err, testPanel)
		}
	}
	app, service, seen, _ := newTestApp(t)
	if err := app.OpenPanel("sideboard"); !errors.Is(err, ribbon.ErrUnknownChoice) || app.panelOpen.Load() || len(seen.placed) != 0 || len(service.at) != 0 {
		t.Errorf("an unknown panel answered %v, open %v, placed %d times", err, app.panelOpen.Load(), len(seen.placed))
	}
}
