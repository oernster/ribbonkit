package desktop

import (
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
)

func TestTheChosenIdentifierNamesItsItem(t *testing.T) {
	t.Parallel()
	items := []menus.Item{{Action: menus.Hide}, {Action: menus.Exit}}
	cases := map[int]struct {
		action menus.Action
		ok     bool
	}{
		0:              {"", false},
		menuIDBase:     {menus.Hide, true},
		menuIDBase + 1: {menus.Exit, true},
		menuIDBase + 2: {"", false},
	}
	for id, want := range cases {
		if action, ok := chosenAction(items, id); action != want.action || ok != want.ok {
			t.Errorf("id %d: got %q %v", id, action, ok)
		}
	}
}

// procGetMenuState reads an item's state back from Windows, which only the test needs.
var procGetMenuState = user32.NewProc("GetMenuState")

// FR-408, FR-412: a disabled item is greyed in the menu Windows builds, by Windows' own reading of it;
// the item beside it is not.
func TestADisabledItemIsGreyedByWindows(t *testing.T) {
	t.Parallel()
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		t.Fatal("Windows made no menu")
	}
	defer func() { _, _, _ = procDestroyMenu.Call(menu) }()
	next := 0
	fill(menu, []menus.Item{{Action: menus.LeftEdge, Label: "Left"}, {Action: menus.RightEdge, Label: "Right", Disabled: true}}, &next)
	greyed := func(id int) bool {
		state, _, _ := procGetMenuState.Call(menu, uintptr(menuIDBase+id), mfByCommand)
		return state&mfGrayed != 0
	}
	if greyed(0) || !greyed(1) {
		t.Errorf("left greyed %v, right %v; want false then true", greyed(0), greyed(1))
	}
}

// FR-401: the threshold is Windows' own drag rectangle at 100 percent, which is never zero.
func TestTheDragThresholdIsWindowsOwn(t *testing.T) {
	t.Parallel()
	got := DragThreshold()
	if got.Width <= 0 || got.Height <= 0 {
		t.Errorf("got %+v", got)
	}
	t.Logf("drag threshold %+v DIP", got)
}
