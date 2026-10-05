package window

// The facade's production calls into Wails. App holds each as a field so the facade's tests can
// stand in for Wails; newApp points the fields here.

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// wailsContext answers the context Wails handed startup; nil before startup.
func (a *Window) wailsContext() context.Context {
	if held := a.ctx.Load(); held != nil {
		return *held
	}
	return nil
}

// emitToWails sends an event to the page. Before startup there is no context to send it into, so an
// event raised then is dropped.
func (a *Window) emitToWails(event string, data ...any) {
	if ctx := a.wailsContext(); ctx != nil {
		runtime.EventsEmit(ctx, event, data...)
	}
}

func (a *Window) showInWails() { runtime.WindowShow(a.wailsContext()) }

func (a *Window) hideInWails() { runtime.WindowHide(a.wailsContext()) }

func (a *Window) quitWails() { runtime.Quit(a.wailsContext()) }

func (a *Window) setOnTopInWails(on bool) { runtime.WindowSetAlwaysOnTop(a.wailsContext(), on) }

// backgroundInWails paints the window and its web view with a colour, opaque or clear (FR-622).
func (a *Window) backgroundInWails(red, green, blue, alpha uint8) {
	runtime.WindowSetBackgroundColour(a.wailsContext(), red, green, blue, alpha)
}
