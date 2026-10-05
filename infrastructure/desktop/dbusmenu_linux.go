package desktop

import (
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/oernster/ribbonkit/application/menus"
)

// The com.canonical.dbusmenu vocabulary the tray's menu is written in.
const (
	menuRootID          int32 = 0
	menuChildrenDisplay       = "children-display"
	menuSubmenu               = "submenu"
	menuLabel                 = "label"
	menuType                  = "type"
	menuSeparator             = "separator"
	menuToggleType            = "toggle-type"
	menuCheckmark             = "checkmark"
	menuToggleState           = "toggle-state"
	menuClicked               = "clicked"
	menuEnabled               = "enabled"
	menuToggleOn        int32 = 1
	menuToggleOff       int32 = 0
)

// menuNode is one entry of a dbusmenu layout: its id, its properties and its children, each child
// a menuNode held in a variant, as the specification's (ia{sv}av) has it.
type menuNode struct {
	ID         int32
	Properties map[string]dbus.Variant
	Children   []dbus.Variant
}

// layoutOf answers items as a dbusmenu layout under the root, with the action each chosen id
// stands for. Ids are given out depth first from one. A separator and a submenu take an id too, as
// every entry must; neither carries an action.
func layoutOf(items []menus.Item) (menuNode, map[int32]menus.Action) {
	actions := map[int32]menus.Action{}
	next := menuRootID
	root := menuNode{ID: menuRootID, Properties: map[string]dbus.Variant{menuChildrenDisplay: dbus.MakeVariant(menuSubmenu)}}
	root.Children = childrenOf(items, &next, actions)
	return root, actions
}

// childrenOf answers items as layout children, numbering them on from next.
func childrenOf(items []menus.Item, next *int32, actions map[int32]menus.Action) []dbus.Variant {
	var children []dbus.Variant
	for _, item := range items {
		if separatedBefore(item) {
			*next++
			children = append(children, dbus.MakeVariant(menuNode{
				ID: *next, Properties: map[string]dbus.Variant{menuType: dbus.MakeVariant(menuSeparator)},
			}))
		}
		*next++
		node := menuNode{ID: *next, Properties: map[string]dbus.Variant{menuLabel: dbus.MakeVariant(escapeLabel(item.Label))}}
		if len(item.Children) > 0 {
			node.Properties[menuChildrenDisplay] = dbus.MakeVariant(menuSubmenu)
			node.Children = childrenOf(item.Children, next, actions)
		} else {
			actions[node.ID] = item.Action
			if item.Disabled {
				node.Properties[menuEnabled] = dbus.MakeVariant(false)
			}
			if item.Checkable {
				node.Properties[menuToggleType] = dbus.MakeVariant(menuCheckmark)
				node.Properties[menuToggleState] = dbus.MakeVariant(toggleState(item.Checked))
			}
		}
		children = append(children, dbus.MakeVariant(node))
	}
	return children
}

// escapeLabel doubles every underscore, which dbusmenu would otherwise read as marking the next
// letter as the item's access key.
func escapeLabel(label string) string { return strings.ReplaceAll(label, "_", "__") }

// toggleState answers the dbusmenu toggle state for checked.
func toggleState(checked bool) int32 {
	if checked {
		return menuToggleOn
	}
	return menuToggleOff
}

// find answers the node with id beneath and including node; false when there is none.
func find(node menuNode, id int32) (menuNode, bool) {
	if node.ID == id {
		return node, true
	}
	for _, child := range node.Children {
		if found, ok := find(child.Value().(menuNode), id); ok {
			return found, true
		}
	}
	return menuNode{}, false
}
