package menus

import (
	"reflect"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// FR-408: each Position item names its edge and nothing else does.
func TestEachPositionActionNamesItsEdge(t *testing.T) {
	t.Parallel()
	edges := map[Action]placement.Edge{LeftEdge: placement.Left, RightEdge: placement.Right, TopEdge: placement.Top, BottomEdge: placement.Bottom}
	for action, edge := range edges {
		if got, ok := EdgeOf(action); !ok || got != edge {
			t.Errorf("%s: got %s, %v; want %s", action, got, ok, edge)
		}
	}
	if _, ok := EdgeOf(Settings); ok {
		t.Error("Settings was taken for an edge")
	}
}

// FR-408, FR-412: Offered greys exactly the Position items whose edge would not move the ribbon
// (submenus included); the menu it was given is left as it was.
func TestOfferedGreysOnlyThePositionItemsThatWouldNotMove(t *testing.T) {
	t.Parallel()
	given := []Item{SettingsItem(), PositionItem(ribbon.Vertical), ExitItem()}
	offered := Offered(given, func(edge placement.Edge) bool { return edge != placement.Right })
	greyed := map[Action]bool{}
	var walk func([]Item)
	walk = func(items []Item) {
		for _, item := range items {
			if item.Disabled {
				greyed[item.Action] = true
			}
			walk(item.Children)
		}
	}
	walk(offered)
	if !reflect.DeepEqual(greyed, map[Action]bool{RightEdge: true}) {
		t.Errorf("greyed %v, want Centre on right edge alone", greyed)
	}
	if given[1].Children[1].Disabled {
		t.Error("Offered changed the menu it was given")
	}
	if Offered(nil, nil) != nil {
		t.Error("no items offered some")
	}
}
