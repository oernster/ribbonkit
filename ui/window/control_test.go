package window

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// controlOver answers the Control over a test's window.
func controlOver(app *Window) *Control { return &Control{window: app} }

// The Control's calls reach the window they hold: a fit, a redraw, a report, a panel shown, the page
// counted measured and the tray counted up.
func TestTheControlReachesItsWindow(t *testing.T) {
	app, service, seen, log := newTestApp(t)
	control := controlOver(app)
	if err := control.Refitted(errPlanted); err != errPlanted {
		t.Errorf("Refitted answered %v, want the error it was handed", err)
	}
	control.ContentChanged()
	if !slices.Equal(service.calls, []string{"Rearrange", "Rearrange"}) {
		t.Errorf("the service heard %v, want the ribbon fitted once for each call", service.calls)
	}
	control.Redraw()
	if err := control.Redrawn(errPlanted); err != errPlanted || len(seen.events) != 2 {
		t.Errorf("Redrawn answered %v after %d events, want the error and two redraws", err, len(seen.events))
	}
	control.Report("doing the planted thing", errPlanted)
	if !strings.Contains(log.String(), "doing the planted thing: planted failure") {
		t.Errorf("the report did not reach the log: %q", log)
	}
	control.ShowPanel("add-clock")
	if seen.shown != 1 || !seen.sawEvent(eventOpenPanel, "add-clock") || !control.Visible() {
		t.Errorf("shown %d times, sent %v; want the ribbon shown and the panel asked for", seen.shown, seen.events)
	}
	control.TrayStarted()
	if !app.trayUp.Load() {
		t.Error("the tray was not counted up")
	}
	control.PageMeasured()
	if !app.launch.measured.Load() {
		t.Error("the page was not counted measured")
	}
	if control.Shown() != app.shown() {
		t.Error("the Control's reading differs from the window's")
	}
}

// ExitWhen ends the application on a signal from outside, through the Control as directly.
func TestTheControlExitsOnASignal(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	signals := make(chan os.Signal, 1)
	signals <- os.Interrupt
	controlOver(app).ExitWhen(signals)
	if seen.quits != 1 {
		t.Errorf("quit %d times, want once", seen.quits)
	}
}
