//go:build linux || darwin

package platform

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/oernster/ribbonkit/infrastructure/desktop"
)

// Prepare gives the desktop icon, a PNG, for the tray on Linux and the menu bar on macOS, then has a
// request to end from outside reach exitWhen, the window's ExitWhen, on a goroutine of its own. A
// shutdown or restart about to begin is such a request too (watchShutdown); its failures go to the
// log, which standard error is written to.
func Prepare(desk *desktop.Desktop, icon []byte, exitWhen func(<-chan os.Signal)) {
	desk.UseIcon(icon)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	watchShutdown(signals, os.Stderr)
	go exitWhen(signals)
}
