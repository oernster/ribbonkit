package desktop

import "github.com/oernster/ribbonkit/application/shell"

// eventBuffer is how many events may wait unread before the next is dropped rather than block the
// thread the desktop calls in on.
const eventBuffer = 32

// Event, EventKind and the kinds are the shell port's own, named here too so the platform halves,
// several of which compile only on their own platform, read them unqualified. Each is an alias: one
// type, one value.
type (
	Event     = shell.Event
	EventKind = shell.EventKind
)

// The events the desktop reports; shell documents each.
const (
	EventMenu           = shell.EventMenu
	EventIconClicked    = shell.EventIconClicked
	EventMoveEnded      = shell.EventMoveEnded
	EventDisplayChanged = shell.EventDisplayChanged
	EventTimeChanged    = shell.EventTimeChanged
	EventResumed        = shell.EventResumed
	EventPointerArrived = shell.EventPointerArrived
	EventPointerLeft    = shell.EventPointerLeft
	EventMenuClosed     = shell.EventMenuClosed
)
