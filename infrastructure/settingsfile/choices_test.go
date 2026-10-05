package settingsfile

import (
	"reflect"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// every answers choices with every one away from its default.
func every() ribbon.Choices {
	return ribbon.Choices{
		Colour: ribbon.Sunset, Orientation: ribbon.Horizontal, Theme: ribbon.Dark, AlwaysOnTop: true,
		SkippedUpdate: "v2.1.0",
		Placement: &placement.Stored{
			Device: `\\.\DISPLAY2`, Work: placement.Rect{Left: 1920, Top: 8, Right: 4480, Bottom: 1392},
			DPI: 144, Offset: placement.Point{X: 180, Y: -4},
		},
		LastEdge:    &placement.Against{Device: `\\.\DISPLAY2`, Edge: placement.Left},
		Opacity:     55,
		Scale:       150,
		PullOutSide: placement.Right,
	}
}

// Every choice is written and read back as it was.
func TestEveryChoiceRoundTrips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	saved := codec.Defaults()
	saved.Choices = every()
	if err := newStore(dir).Save(saved); err != nil {
		t.Fatal(err)
	}
	loaded, _, err := newStore(dir).Load()
	if err != nil || !reflect.DeepEqual(loaded.Choices, every()) {
		t.Errorf("read %+v (%v), want %+v", loaded.Choices, err, every())
	}
}

// A placement or remembered edge that is missing, malformed or names no display is none; a bad
// value leaves its default and the rest load.
func TestABadChoiceLeavesItsDefault(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"missing":    `{"theme": "dark"}`,
		"not a list": `{"theme": "dark", "lastEdge": [1], "placement": [1]}`,
		"no display": `{"theme": "dark", "lastEdge": {"edge": "left"}, "placement": {"device": ""}}`,
		"bad values": `{"theme": "dark", "alwaysOnTop": "yes", "placement": {"device": 3}, "opacity": "most"}`,
	} {
		object, ok := parse([]byte(body))
		if !ok {
			t.Fatalf("%s: not an object", name)
		}
		got := ribbon.Defaults()
		ReadChoices(object, &got)
		want := ribbon.Defaults()
		want.Theme = ribbon.Dark
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: read %+v", name, got)
		}
	}
}
