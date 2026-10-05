package desktop

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/identity"

	"golang.org/x/sys/windows"
)

// Desktop is the hidden window that owns the tray icon and hears the desktop's broadcasts. Every
// Win32 handle it holds belongs to one locked thread; other goroutines only post to it.
type Desktop struct {
	app    identity.App
	menu   func() []menus.Item
	events chan Event
	log    io.Writer

	ribbon  atomic.Uintptr
	window  windows.HWND
	posted  atomic.Uintptr
	icon    windows.Handle
	hook    uintptr
	created uint32
	pending pendingMenu
	pointer pointerTracker

	started sync.Once
	stopped sync.Once
	ready   chan error
}

// New answers a desktop for app whose tray menu is menu, called each time the menu opens, reporting
// failures on its own thread to log.
func New(app identity.App, menu func() []menus.Item, log io.Writer) *Desktop {
	return &Desktop{app: app, menu: menu, events: make(chan Event, eventBuffer), log: log, ready: make(chan error, 1)}
}

// className is the hidden window's class, unique to the application.
func (d *Desktop) className() string { return d.app.Name + "Desktop" }

// Events yields what happened. The channel is closed when the desktop stops.
func (d *Desktop) Events() <-chan Event { return d.events }

// Watch names the ribbon's window, so the end of its moves is reported.
func (d *Desktop) Watch(ribbon Window) { d.ribbon.Store(uintptr(ribbon)) }

// Start shows the tray icon and runs the message loop on its own locked thread, returning once the
// icon is there or with the reason it is not.
func (d *Desktop) Start() error {
	var err error
	d.started.Do(func() {
		go d.run()
		err = <-d.ready
	})
	return err
}

// Stop removes the icon and ends the loop.
func (d *Desktop) Stop() {
	d.stopped.Do(func() {
		if window := d.posted.Load(); window != 0 {
			_, _, _ = procPostMessage.Call(window, wmClose, 0, 0)
		}
	})
}

func (d *Desktop) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(d.events)
	if err := d.create(); err != nil {
		d.ready <- err
		return
	}
	d.ready <- nil
	d.pump()
	d.destroy()
}

// create registers the class, makes the hidden window, adds the icon and hooks the end of moves.
func (d *Desktop) create() error {
	instance, _, _ := procGetModuleHandle.Call(0)
	classText, err := windows.UTF16PtrFromString(d.className())
	if err != nil {
		return fmt.Errorf("encoding the desktop window class: %w", err)
	}
	name := uintptr(unsafe.Pointer(classText))
	class := wndClassEx{lpfnWndProc: windows.NewCallback(d.windowProc), hInstance: windows.Handle(instance), lpszClassName: classText}
	class.cbSize = uint32(unsafe.Sizeof(class))
	if ret, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&class))); ret == 0 {
		return fmt.Errorf("registering the desktop window class: %w", err)
	}
	// A top-level window with no parent, never shown: unlike a message-only window, it hears the
	// display, time and power broadcasts.
	handle, _, err := procCreateWindowEx.Call(0, name, name, 0, 0, 0, 0, 0, 0, 0, instance, 0)
	if handle == 0 {
		return fmt.Errorf("creating the desktop window: %w", err)
	}
	d.window = windows.HWND(handle)
	created, _, _ := procRegisterWindowMessage.Call(utf16Pointer(taskbarCreatedMsg))
	d.created = uint32(created)
	d.icon = ownIcon()
	if err := d.addIcon(); err != nil {
		_, _, _ = procDestroyWindow.Call(handle)
		return err
	}
	pid, _, _ := procGetCurrentProcessID.Call()
	d.hook, _, _ = procSetWinEventHook.Call(eventMoveSizeEnd, eventMoveSizeEnd, 0,
		windows.NewCallback(d.moveEnded), pid, 0, winEventOutOfCtx)
	d.posted.Store(handle)
	return nil
}

// addIcon adds the tray icon with its tooltip (FR-501).
func (d *Desktop) addIcon() error {
	data := d.iconData()
	if ret, _, err := procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&data))); ret == 0 {
		return fmt.Errorf("adding the tray icon: %w", err)
	}
	return nil
}

func (d *Desktop) iconData() notifyIconData {
	data := notifyIconData{hWnd: d.window, uID: trayIconID, uFlags: nifMessage | nifIcon | nifTip, uCallbackMessage: wmTrayCallback, hIcon: d.icon}
	data.cbSize = uint32(unsafe.Sizeof(data))
	tip, _ := windows.UTF16FromString(d.app.Name)
	copy(data.szTip[:tipLength-1], tip)
	return data
}

func (d *Desktop) pump() {
	var message msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) <= 0 {
			return
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func (d *Desktop) destroy() {
	d.posted.Store(0)
	if d.hook != 0 {
		_, _, _ = procUnhookWinEvent.Call(d.hook)
	}
	data := d.iconData()
	_, _, _ = procShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
	_, _, _ = procDestroyWindow.Call(uintptr(d.window))
}

// windowProc answers the hidden window's messages. A panic here would end the application from a
// thread Windows called in on, so it is recovered and written to the log instead.
func (d *Desktop) windowProc(hwnd windows.HWND, message uint32, wParam, lParam uintptr) (result uintptr) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(d.log, "desktop: recovered from %v\n", failure)
			result = 0
		}
	}()
	switch {
	case message == wmTrayCallback && lParam == wmRButtonUp:
		d.track(d.menu())
	case message == wmShowMenu:
		d.showPending()
	case message == wmTrayCallback && lParam == wmLButtonUp:
		d.send(Event{Kind: EventIconClicked})
	case message == wmDisplayChange:
		d.send(Event{Kind: EventDisplayChanged})
	case message == wmTimeChange:
		d.send(Event{Kind: EventTimeChanged})
	case message == wmPowerBroadcast && (wParam == pbtResumeAuto || wParam == pbtResumeSuspend):
		d.send(Event{Kind: EventResumed})
	case d.created != 0 && message == d.created:
		// Explorer restarted, taking the icon with it.
		_ = d.addIcon()
	case message == wmClose:
		_, _, _ = procDestroyWindow.Call(uintptr(hwnd))
	case message == wmDestroy:
		_, _, _ = procPostQuitMessage.Call(0)
	default:
		result, _, _ = procDefWindowProc.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
		return result
	}
	return 0
}

// moveEnded is the WinEvent callback for the end of a move anywhere in this process.
func (d *Desktop) moveEnded(_, _, hwnd, _, _, _, _ uintptr) uintptr {
	if hwnd != 0 && hwnd == d.ribbon.Load() {
		d.send(Event{Kind: EventMoveEnded})
	}
	return 0
}

// send offers an event without blocking the thread Windows called in on.
func (d *Desktop) send(event Event) {
	select {
	case d.events <- event:
	default:
		fmt.Fprintf(d.log, "desktop: event %d dropped, nothing is reading\n", event.Kind)
	}
}

// ownIcon loads the icon built into this executable; the shell's application icon when it has
// none, which is how an unpackaged build runs.
func ownIcon() windows.Handle {
	if executable, err := os.Executable(); err == nil {
		var large, small windows.Handle
		ret, _, _ := procExtractIconEx.Call(utf16Pointer(executable), 0,
			uintptr(unsafe.Pointer(&large)), uintptr(unsafe.Pointer(&small)), 1)
		if ret > 0 && small != 0 {
			return small
		}
	}
	handle, _, _ := procLoadIcon.Call(0, idiApplication)
	return windows.Handle(handle)
}
