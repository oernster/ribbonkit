package window

import (
	"errors"
	"io"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/shell"
	"github.com/oernster/ribbonkit/domain/placement"
)

// recordingDesktop is a shell.Desktop that records each call by name with the window it was given.
type recordingDesktop struct {
	calls   []string
	windows []shell.Window
}

func (d *recordingDesktop) saw(call string, ribbon shell.Window) {
	d.calls = append(d.calls, call)
	d.windows = append(d.windows, ribbon)
}

func (d *recordingDesktop) Events() <-chan shell.Event { return nil }
func (d *recordingDesktop) Watch(ribbon shell.Window)  { d.saw("Watch", ribbon) }
func (d *recordingDesktop) Stop()                      { d.saw("Stop", 0) }
func (d *recordingDesktop) ShowMenu([]menus.Item)      { d.saw("ShowMenu", 0) }
func (d *recordingDesktop) TrackPointer(ribbon shell.Window, _ bool) {
	d.saw("TrackPointer", ribbon)
}
func (d *recordingDesktop) FindRibbon(string, string) (shell.Window, error) { return 0, errPortTest }
func (d *recordingDesktop) HideFromTaskbar(ribbon shell.Window) error {
	d.saw("HideFromTaskbar", ribbon)
	return nil
}
func (d *recordingDesktop) KeepOnDisplays(ribbon shell.Window, _ io.Writer) error {
	d.saw("KeepOnDisplays", ribbon)
	return nil
}
func (d *recordingDesktop) Place(ribbon shell.Window, _ placement.Point, _ placement.Size) error {
	d.saw("Place", ribbon)
	return nil
}
func (d *recordingDesktop) Position(ribbon shell.Window) (placement.Point, error) {
	d.saw("Position", ribbon)
	return placement.Point{}, nil
}
func (d *recordingDesktop) Shape(ribbon shell.Window, _ []placement.Rect) error {
	d.saw("Shape", ribbon)
	return nil
}
func (d *recordingDesktop) SetTabFrame(ribbon shell.Window, _ bool) error {
	d.saw("SetTabFrame", ribbon)
	return nil
}
func (d *recordingDesktop) Cursor() (placement.Point, bool) {
	d.saw("Cursor", 0)
	return placement.Point{}, false
}
func (d *recordingDesktop) DragThreshold() placement.Size {
	d.saw("DragThreshold", 0)
	return placement.Size{}
}
func (d *recordingDesktop) ToolkitScale() int {
	d.saw("ToolkitScale", 0)
	return testUnscaled
}
func (d *recordingDesktop) PixelsPerDIP(pageRatio float64, _ int) float64 {
	d.saw("PixelsPerDIP", 0)
	return pageRatio
}
func (d *recordingDesktop) OpenInBrowser(string) error {
	d.saw("OpenInBrowser", 0)
	return nil
}

// errPortTest is what the recording desktop's FindRibbon answers, so a test can tell it was asked.
var errPortTest = errors.New("the recording desktop finds no window")

// Every call the window makes into the desktop goes through the port it was built with, carrying the
// ribbon's own window where the call is about it.
func TestTheWindowReachesTheDesktopThroughItsPort(t *testing.T) {
	t.Parallel()
	desk := &recordingDesktop{}
	app, _ := New(Config{Service: &scriptedService{arrangement: testArrange}, Desktop: desk, Log: io.Discard})
	app.ribbon = testRibbon
	_ = app.browse("https://example.invalid")
	app.showMenu(nil)
	_, _ = app.position()
	_ = app.place(placement.Point{}, placement.Size{})
	_ = app.shape(nil)
	_ = app.tabFrame(true)
	app.watchPointer(true)
	_ = app.toolkitScale()
	_ = app.perDIPOf(1, testUnscaled)
	_ = app.dragThreshold()
	_, _ = app.cursor()
	want := []string{
		"OpenInBrowser", "ShowMenu", "Position", "Place", "Shape", "SetTabFrame", "TrackPointer",
		"ToolkitScale", "PixelsPerDIP", "DragThreshold", "Cursor",
	}
	if !slices.Equal(desk.calls, want) {
		t.Fatalf("the window called %v, want %v", desk.calls, want)
	}
	for index, call := range desk.calls {
		about := slices.Contains([]string{"Position", "Place", "Shape", "SetTabFrame", "TrackPointer"}, call)
		if about && desk.windows[index] != testRibbon {
			t.Errorf("%s was given window %d, want the ribbon's", call, desk.windows[index])
		}
	}
}
