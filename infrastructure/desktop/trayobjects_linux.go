package desktop

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

// The objects the tray host calls on the bus. The method names and arguments are the
// specifications'; the bus answers a call only to a method whose last result is a *dbus.Error.

// errUnknownMenuItem is answered for an id no item of the menu carries.
var errUnknownMenuItem = errors.New("no menu item has that id")

// trayItem answers org.kde.StatusNotifierItem.
type trayItem struct{ tray *tray }

// Activate is a primary click on the icon: it shows or hides the ribbon (FR-503).
func (i trayItem) Activate(_, _ int32) *dbus.Error {
	i.tray.desktop.send(Event{Kind: EventIconClicked})
	return nil
}

// SecondaryActivate is a middle click, which does nothing.
func (trayItem) SecondaryActivate(_, _ int32) *dbus.Error { return nil }

// ContextMenu does nothing: the icon names its menu, which the host opens itself.
func (trayItem) ContextMenu(_, _ int32) *dbus.Error { return nil }

// Scroll does nothing.
func (trayItem) Scroll(int32, string) *dbus.Error { return nil }

// trayMenu answers com.canonical.dbusmenu.
type trayMenu struct{ tray *tray }

// groupProperties is one entry of GetGroupProperties' answer.
type groupProperties struct {
	ID         int32
	Properties map[string]dbus.Variant
}

// menuEvent is one entry of EventGroup's argument.
type menuEvent struct {
	ID        int32
	EventID   string
	Data      dbus.Variant
	Timestamp uint32
}

// GetLayout answers the menu beneath parent. Depth and the property names asked for are not
// honoured: the whole subtree with every property is always a correct answer, only a larger one.
func (m trayMenu) GetLayout(parent, _ int32, _ []string) (uint32, menuNode, *dbus.Error) {
	layout, revision := m.tray.current()
	if node, ok := find(layout, parent); ok {
		return revision, node, nil
	}
	return revision, menuNode{}, dbus.MakeFailedError(errUnknownMenuItem)
}

// GetGroupProperties answers the properties of each item named.
func (m trayMenu) GetGroupProperties(ids []int32, _ []string) ([]groupProperties, *dbus.Error) {
	layout, _ := m.tray.current()
	var out []groupProperties
	for _, id := range ids {
		if node, ok := find(layout, id); ok {
			out = append(out, groupProperties{ID: id, Properties: node.Properties})
		}
	}
	return out, nil
}

// GetProperty answers one property of one item.
func (m trayMenu) GetProperty(id int32, name string) (dbus.Variant, *dbus.Error) {
	layout, _ := m.tray.current()
	if node, ok := find(layout, id); ok {
		if value, ok := node.Properties[name]; ok {
			return value, nil
		}
	}
	return dbus.Variant{}, dbus.MakeFailedError(errUnknownMenuItem)
}

// Event hears something done to an item; a click chooses it.
func (m trayMenu) Event(id int32, eventID string, _ dbus.Variant, _ uint32) *dbus.Error {
	if eventID == menuClicked {
		m.tray.chosen(id)
	}
	return nil
}

// EventGroup hears several events at once.
func (m trayMenu) EventGroup(events []menuEvent) ([]int32, *dbus.Error) {
	for _, event := range events {
		_ = m.Event(event.ID, event.EventID, event.Data, event.Timestamp)
	}
	return []int32{}, nil
}

// AboutToShow rebuilds the menu before the host shows it, answering whether it changed so the host
// fetches it again; a change is also announced, for a host that waits for the signal instead.
func (m trayMenu) AboutToShow(int32) (bool, *dbus.Error) {
	changed := m.tray.refresh()
	if changed {
		_, revision := m.tray.current()
		_ = m.tray.conn.Emit(menuPath, layoutUpdated, revision, menuRootID)
	}
	return changed, nil
}

// AboutToShowGroup is AboutToShow for several menus, which are all the one menu here.
func (m trayMenu) AboutToShowGroup(ids []int32) ([]int32, []int32, *dbus.Error) {
	if changed, _ := m.AboutToShow(menuRootID); changed {
		return ids, []int32{}, nil
	}
	return []int32{}, []int32{}, nil
}
