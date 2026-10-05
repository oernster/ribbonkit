// Package menus is the menu model the tray and the ribbon's right-click menu are built from: an item,
// what it does and the actions every ribbon offers. An application adds actions of its own of the
// same type; the desktop shows any of them and reports which was chosen. FR numbers are TimeRibbon's.
package menus

import "github.com/oernster/ribbonkit/domain/placement"

// Action names what a menu item does. The window and the tray act on it.
type Action string

// The actions every ribbon's menus offer.
const (
	Show        Action = "show"
	Hide        Action = "hide"
	Settings    Action = "settings"
	AlwaysOnTop Action = "always-on-top"
	Pin         Action = "pin"
	About       Action = "about"
	Licence     Action = "licence"
	Updates     Action = "check-updates"
	Exit        Action = "exit"
	LeftEdge    Action = "left-edge"
	RightEdge   Action = "right-edge"
	TopEdge     Action = "top-edge"
	BottomEdge  Action = "bottom-edge"
)

// edgeActions maps each Position item to the edge it puts the ribbon against (FR-408).
var edgeActions = map[Action]placement.Edge{
	LeftEdge:   placement.Left,
	RightEdge:  placement.Right,
	TopEdge:    placement.Top,
	BottomEdge: placement.Bottom,
}

// EdgeOf answers the edge a Position item puts the ribbon against; false for any other action.
func EdgeOf(action Action) (placement.Edge, bool) {
	edge, ok := edgeActions[action]
	return edge, ok
}

// Offered answers items as a menu offers them: each Position item disabled where moves answers false
// for its edge, so no menu offers a press that would leave the ribbon where it stands (FR-408,
// FR-412). Nothing else changes; items itself is left as it was.
func Offered(items []Item, moves func(placement.Edge) bool) []Item {
	if items == nil {
		return nil
	}
	offered := make([]Item, len(items))
	for index, item := range items {
		if edge, ok := EdgeOf(item.Action); ok {
			item.Disabled = !moves(edge)
		}
		item.Children = Offered(item.Children, moves)
		offered[index] = item
	}
	return offered
}

// Item is one entry of a menu.
type Item struct {
	Action Action
	Label  string
	// Checkable items show Checked beside their label.
	Checkable bool
	Checked   bool
	// Disabled items are shown greyed and cannot be chosen, as a Position item is where pressing it
	// would leave the ribbon where it stands (Offered).
	Disabled bool
	// Children makes the item a submenu holding them; such an item has no action of its own.
	Children []Item
}
