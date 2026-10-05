package menus

import (
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
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
