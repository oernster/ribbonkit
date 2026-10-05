package desktop

import (
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
)

// FR-508: a submenu takes no number of its own; its items are numbered after every item before it
// and before every item after it.
func TestASubmenuIsNumberedAfterEveryItemBeforeIt(t *testing.T) {
	t.Parallel()
	items := []menus.Item{
		{Action: menus.Settings},
		{Label: "Help", Children: []menus.Item{{Action: menus.About}, {Action: menus.Licence}}},
		{Action: menus.Exit},
	}
	want := []menus.Action{menus.Settings, menus.About, menus.Licence, menus.Exit}
	for index, action := range want {
		if got, ok := actionAt(items, index); !ok || got != action {
			t.Errorf("index %d: got %q %v, want %q", index, got, ok, action)
		}
	}
	for _, index := range []int{-1, len(want)} {
		if _, ok := actionAt(items, index); ok {
			t.Errorf("index %d named an item", index)
		}
	}
}
