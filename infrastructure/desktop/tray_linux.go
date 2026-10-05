package desktop

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"

	"github.com/oernster/ribbonkit/application/menus"
)

// The tray on Linux is a StatusNotifierItem with a com.canonical.dbusmenu menu, the standard the
// desktop's tray host reads over the session bus; on Ubuntu's GNOME the host is the AppIndicator
// extension (measured 2026-09-28: org.kde.StatusNotifierWatcher owned by gnome-shell).
const (
	itemInterface    = "org.kde.StatusNotifierItem"
	itemPath         = dbus.ObjectPath("/StatusNotifierItem")
	menuInterface    = "com.canonical.dbusmenu"
	menuPath         = dbus.ObjectPath("/MenuBar")
	watcherName      = "org.kde.StatusNotifierWatcher"
	watcherPath      = dbus.ObjectPath("/StatusNotifierWatcher")
	watcherRegister  = watcherName + ".RegisterStatusNotifierItem"
	layoutUpdated    = menuInterface + ".LayoutUpdated"
	ownerChanged     = "NameOwnerChanged"
	busInterface     = "org.freedesktop.DBus"
	menuVersion      = uint32(3)
	itemCategory     = "ApplicationStatus"
	itemStatusActive = "Active"
	menuStatusNormal = "normal"
	textLeftToRight  = "ltr"
	signalBuffer     = 8
)

// errNoTrayHost is answered when nothing on the session bus hosts tray icons.
var errNoTrayHost = errors.New("no tray host is running on this desktop")

// tooltip is the specification's tooltip: an icon name, icon images, a title and a text.
type tooltip struct {
	IconName string
	Pixmaps  []pixmap
	Title    string
	Text     string
}

// tray is the ribbon's icon in the desktop's tray and the menu it opens. The menu is built from the
// desktop's menu function whenever the host is about to show it, so it always says what is true now.
type tray struct {
	desktop *Desktop
	conn    *dbus.Conn

	guard    sync.Mutex
	revision uint32
	layout   menuNode
	actions  map[int32]menus.Action
}

// startTray puts the icon in the tray, answering why it could not.
func startTray(d *Desktop, icon []byte) (*tray, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("reaching the session bus: %w", err)
	}
	t := &tray{desktop: d, conn: conn}
	t.refresh()
	if err := t.export(icon); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := t.register(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	go t.followHost()
	return t, nil
}

// export offers the icon and its menu on the tray's connection.
func (t *tray) export(icon []byte) error {
	pixmaps := []pixmap{}
	if image, err := pixmapOf(icon, trayIconSize); err == nil {
		pixmaps = append(pixmaps, image)
	} else {
		fmt.Fprintf(t.desktop.log, "desktop: %v\n", err)
	}
	if err := t.conn.Export(trayItem{t}, itemPath, itemInterface); err != nil {
		return fmt.Errorf("offering the tray icon: %w", err)
	}
	if err := t.conn.Export(trayMenu{t}, menuPath, menuInterface); err != nil {
		return fmt.Errorf("offering the tray menu: %w", err)
	}
	fixed := func(value any) *prop.Prop { return &prop.Prop{Value: value, Emit: prop.EmitTrue} }
	if _, err := prop.Export(t.conn, itemPath, prop.Map{itemInterface: {
		"Category": fixed(itemCategory), "Id": fixed(t.desktop.app.AppID), "Title": fixed(t.desktop.app.Name),
		"Status": fixed(itemStatusActive), "WindowId": fixed(int32(0)), "IconName": fixed(""),
		"IconPixmap": fixed(pixmaps), "ItemIsMenu": fixed(false), "Menu": fixed(menuPath),
		"ToolTip": fixed(tooltip{Pixmaps: []pixmap{}, Title: t.desktop.app.Name}),
	}}); err != nil {
		return fmt.Errorf("describing the tray icon: %w", err)
	}
	if _, err := prop.Export(t.conn, menuPath, prop.Map{menuInterface: {
		"Version": fixed(menuVersion), "TextDirection": fixed(textLeftToRight),
		"Status": fixed(menuStatusNormal), "IconThemePath": fixed([]string{}),
	}}); err != nil {
		return fmt.Errorf("describing the tray menu: %w", err)
	}
	return nil
}

// register tells the tray host where the icon is: by its object path, which the host joins to the
// connection that called (read in Ubuntu's statusNotifierWatcher.js, 2026-09-28). A bus name of the
// form org.kde.StatusNotifierItem-<pid>-1 would need a sandbox permission to own and would clash
// between Flatpaks, which each see small process ids of their own.
func (t *tray) register() error {
	if err := t.conn.Object(watcherName, watcherPath).Call(watcherRegister, 0, string(itemPath)).Err; err != nil {
		return fmt.Errorf("%w: %v", errNoTrayHost, err)
	}
	return nil
}

// followHost registers the icon again whenever a tray host takes over, as when the desktop's shell
// restarts, until the desktop stops.
func (t *tray) followHost() {
	if err := t.conn.AddMatchSignal(dbus.WithMatchInterface(busInterface), dbus.WithMatchMember(ownerChanged),
		dbus.WithMatchArg(0, watcherName)); err != nil {
		fmt.Fprintf(t.desktop.log, "desktop: following the tray host: %v\n", err)
		return
	}
	signals := make(chan *dbus.Signal, signalBuffer)
	t.conn.Signal(signals)
	defer t.conn.RemoveSignal(signals)
	for {
		select {
		case <-t.desktop.stop:
			return
		case signal, open := <-signals:
			if !open {
				return
			}
			if newOwner, ok := signal.Body[len(signal.Body)-1].(string); ok && newOwner != "" {
				if err := t.register(); err != nil {
					fmt.Fprintf(t.desktop.log, "desktop: %v\n", err)
				}
			}
		}
	}
}

// refresh rebuilds the menu from what is true now, answering whether it changed.
func (t *tray) refresh() bool {
	layout, actions := layoutOf(t.desktop.menu())
	t.guard.Lock()
	defer t.guard.Unlock()
	if reflect.DeepEqual(layout, t.layout) {
		return false
	}
	t.layout, t.actions = layout, actions
	t.revision++
	return true
}

// current answers the menu as last built and its revision.
func (t *tray) current() (menuNode, uint32) {
	t.guard.Lock()
	defer t.guard.Unlock()
	return t.layout, t.revision
}

// chosen reports the item with id as chosen.
func (t *tray) chosen(id int32) {
	t.guard.Lock()
	action, ok := t.actions[id]
	t.guard.Unlock()
	if ok {
		t.desktop.send(Event{Kind: EventMenu, Action: action})
	}
}

// stop takes the icon out of the tray: the host drops an item whose bus name has gone.
func (t *tray) stop() { _ = t.conn.Close() }
