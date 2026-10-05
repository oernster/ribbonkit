package menus

import (
	"reflect"
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// The visibility item names the opposite of what is (FR-502).
func TestVisibilitySaysWhatPressingWillDo(t *testing.T) {
	t.Parallel()
	if got := Visibility(true); !reflect.DeepEqual(got, HideItem()) || got.Label != labelHide {
		t.Errorf("visible: %+v", got)
	}
	if got := Visibility(false); got.Action != Show || got.Label != labelShow {
		t.Errorf("hidden: %+v", got)
	}
}

// The commands carry their actions and words.
func TestTheCommandsCarryTheirActions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		item   Item
		action Action
		label  string
	}{
		{SettingsItem(), Settings, labelSettings},
		{ExitItem(), Exit, labelExit},
		{AlwaysOnTopItem(true), AlwaysOnTop, labelAlwaysOnTop},
		{PinItem(false), Pin, labelPin},
	} {
		if tc.item.Action != tc.action || tc.item.Label != tc.label {
			t.Errorf("%+v", tc.item)
		}
	}
	if !AlwaysOnTopItem(true).Checked || PinItem(false).Checked || !PinItem(true).Checkable {
		t.Error("the ticks do not follow the choices")
	}
	help := HelpItem()
	if help.Label != labelHelp || len(help.Children) != 3 || help.Children[2].Action != Updates {
		t.Errorf("help %+v", help)
	}
}

// A vertical ribbon is offered the left and right edges, a horizontal one the top and bottom (FR-408).
func TestPositionOffersTheEdgesAlongTheRibbon(t *testing.T) {
	t.Parallel()
	actions := func(item Item) []Action {
		var out []Action
		for _, child := range item.Children {
			out = append(out, child.Action)
		}
		return out
	}
	if got := actions(PositionItem(ribbon.Vertical)); !reflect.DeepEqual(got, []Action{LeftEdge, RightEdge}) {
		t.Errorf("vertical %v", got)
	}
	if got := actions(PositionItem(ribbon.Horizontal)); !reflect.DeepEqual(got, []Action{TopEdge, BottomEdge}) {
		t.Errorf("horizontal %v", got)
	}
}

// Every scheme is offered with the current one ticked; each item chooses its scheme (FR-611).
func TestEveryColourIsOfferedAndChosen(t *testing.T) {
	t.Parallel()
	item := ColourItem(ribbon.Ocean)
	if len(item.Children) != len(ribbon.Colours) {
		t.Fatalf("%d colours offered", len(item.Children))
	}
	for _, child := range item.Children {
		colour, ok := ColourOf(child.Action)
		if !ok || child.Checked != (colour == ribbon.Ocean) || child.Label == "" {
			t.Errorf("%+v chooses %q, %v", child, colour, ok)
		}
	}
	for _, action := range []Action{"colour-plaid", "plaid", Exit} {
		if _, ok := ColourOf(action); ok {
			t.Errorf("%q chose a colour", action)
		}
	}
}

// Both orientations are offered with the current one ticked; each item chooses its orientation.
func TestEachOrientationIsOfferedAndChosen(t *testing.T) {
	t.Parallel()
	item := OrientationItem(ribbon.Vertical)
	for _, child := range item.Children {
		orientation, ok := OrientationOf(child.Action)
		if !ok || child.Checked != (orientation == ribbon.Vertical) {
			t.Errorf("%+v chooses %q, %v", child, orientation, ok)
		}
	}
	if _, ok := OrientationOf(Exit); ok {
		t.Error("Exit chose an orientation")
	}
}
