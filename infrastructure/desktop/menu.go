package desktop

import "github.com/oernster/ribbonkit/application/menus"

// numbered answers the action of every item a native menu gives a number, in the order they are
// given out: depth first, so a submenu's items follow every item before it. A submenu takes no
// number of its own (FR-508).
func numbered(items []menus.Item) []menus.Action {
	var out []menus.Action
	for _, item := range items {
		if len(item.Children) > 0 {
			out = append(out, numbered(item.Children)...)
			continue
		}
		out = append(out, item.Action)
	}
	return out
}

// separatedBefore answers whether a native menu draws a separator above item: Exit stands apart
// from everything before it.
func separatedBefore(item menus.Item) bool { return item.Action == menus.Exit }

// actionAt answers the action numbered index; false for a number no item carries.
func actionAt(items []menus.Item, index int) (menus.Action, bool) {
	actions := numbered(items)
	if index < 0 || index >= len(actions) {
		return "", false
	}
	return actions[index], true
}
