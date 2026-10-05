package window

import (
	"context"
	"os"
	"testing"
)

// A signal from outside ends the application even with the tray up, where a close only hides it.
func TestASignalExitsEvenWithTheTrayUp(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.trayUp.Store(true)
	signals := make(chan os.Signal, 1)
	signals <- os.Interrupt
	app.exitWhen(signals)
	if seen.quits != 1 || app.beforeClose(context.Background()) {
		t.Errorf("quit %d times, close held back %v; want the application to end", seen.quits, app.beforeClose(context.Background()))
	}
}

// Signals that stop coming end nothing.
func TestClosedSignalsExitNothing(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	signals := make(chan os.Signal)
	close(signals)
	app.exitWhen(signals)
	if seen.quits != 0 {
		t.Errorf("quit %d times", seen.quits)
	}
}
