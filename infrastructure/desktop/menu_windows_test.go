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

// FR-401: the threshold is Windows' own drag rectangle at 100 percent, which is never zero.
func TestTheDragThresholdIsWindowsOwn(t *testing.T) {
	t.Parallel()
	got := DragThreshold()
	if got.Width <= 0 || got.Height <= 0 {
		t.Errorf("got %+v", got)
	}
	t.Logf("drag threshold %+v DIP", got)
}
