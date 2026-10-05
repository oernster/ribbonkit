package desktop

// The pointer on the ribbon or off it (FR-615, FR-616). Each platform tells it in its own way:
// pointer_poll.go reads it on Windows and macOS, pointer_linux.go is told it by GTK.

// pointerEvent answers the event for the pointer on the ribbon or off it.
func pointerEvent(inside bool) Event {
	if inside {
		return Event{Kind: EventPointerArrived}
	}
	return Event{Kind: EventPointerLeft}
}
