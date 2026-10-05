// Package shell is what a ribbon asks of the desktop shell below its window: the tray icon and its
// menu, the events the desktop reports, the operations on the ribbon's own window and the desktop's
// readings of the pointer and of scale. Desktop is the port; the kit's desktop package implements
// it on each platform and the composition root hands it in.
//
// FR numbers are TimeRibbon's REQUIREMENTS.md, where each rule was first specified.
package shell

import (
	"io"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
)

// Window is the ribbon's window as the platform names it: a window handle on Windows, the ribbon's
// place in the desktop's own list of windows on Linux. Zero is no window, which is what a ribbon
// holds until it has found its own.
type Window uintptr

// EventKind names what happened.
type EventKind int

// The events the desktop reports.
const (
	// EventMenu is a menu item chosen; Event.Action names it.
	EventMenu EventKind = iota
	// EventIconClicked is a left click on the tray icon (FR-503).
	EventIconClicked
	// EventMoveEnded is the end of a move of the ribbon's window (FR-404).
	EventMoveEnded
	// EventDisplayChanged is a change of displays, resolution or arrangement (FR-406).
	EventDisplayChanged
	// EventTimeChanged is a change of the system time or time zone (FR-209).
	EventTimeChanged
	// EventResumed is a resume from sleep (FR-209).
	EventResumed
	// EventPointerArrived is the pointer come onto the ribbon or its tab (FR-615).
	EventPointerArrived
	// EventPointerLeft is the pointer gone off the ribbon or its tab (FR-616).
	EventPointerLeft
	// EventMenuClosed is a popup menu closed, chosen from or not; one open holds the ribbon (FR-616).
	EventMenuClosed
)

// Event is one thing that happened on the desktop.
type Event struct {
	Kind   EventKind
	Action menus.Action
}

// Desktop is the desktop shell as a ribbon sees it.
type Desktop interface {
	// Events answers what the desktop reports, until Stop.
	Events() <-chan Event
	// Watch starts reporting the end of a move of ribbon (FR-404).
	Watch(ribbon Window)
	// Stop ends the reporting and removes the tray icon.
	Stop()
	// ShowMenu shows items as a native menu at the pointer (FR-108).
	ShowMenu(items []menus.Item)
	// TrackPointer starts or stops reporting the pointer coming onto ribbon and going off it.
	TrackPointer(ribbon Window, on bool)

	// FindRibbon answers the ribbon's window by its class and its title.
	FindRibbon(class, name string) (Window, error)
	// HideFromTaskbar takes ribbon off the taskbar (FR-101).
	HideFromTaskbar(ribbon Window) error
	// KeepOnDisplays keeps a drag of ribbon from leaving every display, writing what it did to log.
	KeepOnDisplays(ribbon Window, log io.Writer) error
	// Place puts ribbon at at, size across, in physical pixels.
	Place(ribbon Window, at placement.Point, size placement.Size) error
	// Position answers where ribbon's top-left corner stands, in physical pixels.
	Position(ribbon Window) (placement.Point, error)
	// Shape cuts ribbon to parts, in its own pixels (FR-913).
	Shape(ribbon Window, parts []placement.Rect) error
	// SetTabFrame gives ribbon the tab's frame or gives the toolkit's back.
	SetTabFrame(ribbon Window, tab bool) error

	// Cursor answers where the pointer is, in physical pixels; false where it cannot be read.
	Cursor() (placement.Point, bool)
	// DragThreshold answers how far the pointer moves before a press becomes a drag.
	DragThreshold() placement.Size
	// ToolkitScale answers the toolkit's own window scale, which the page's ratio is divided by.
	ToolkitScale() int
	// PixelsPerDIP answers the window pixels to each of the page's units for the page's ratio.
	PixelsPerDIP(pageRatio float64, toolkitScale int) float64
	// OpenInBrowser hands address to the desktop's browser.
	OpenInBrowser(address string) error
}
