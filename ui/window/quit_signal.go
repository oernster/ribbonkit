package window

import (
	"os"

	"github.com/oernster/ribbonkit/application/menus"
)

// exitWhen ends the application once a signal arrives on signals: a request from outside, as when
// the session ends, which must end it even while the tray icon makes closing the ribbon only hide
// it. Wails turns the same signal into an ordinary close, which beforeClose then holds back
// (measured 2026-09-28: with the tray up, SIGTERM left the ribbon running).
func (a *Window) exitWhen(signals <-chan os.Signal) {
	if _, open := <-signals; open {
		a.act(menus.Exit)
	}
}
