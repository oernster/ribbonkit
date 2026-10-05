package window

// The facade's production calls into Wails. App holds each as a field so the facade's tests can
// stand in for Wails; newApp points the fields here.

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// emitToWails sends an event to the page. Before startup there is no context to send it into, so an
// event raised then is dropped.
func (a *Window) emitToWails(event string, data ...any) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data...)
	}
}

func (a *Window) showInWails() { runtime.WindowShow(a.ctx) }

func (a *Window) hideInWails() { runtime.WindowHide(a.ctx) }

func (a *Window) quitWails() { runtime.Quit(a.ctx) }

func (a *Window) setOnTopInWails(on bool) { runtime.WindowSetAlwaysOnTop(a.ctx, on) }

// backgroundInWails paints the window and its web view with a colour, opaque or clear (FR-622).
func (a *Window) backgroundInWails(red, green, blue, alpha uint8) {
	runtime.WindowSetBackgroundColour(a.ctx, red, green, blue, alpha)
}
