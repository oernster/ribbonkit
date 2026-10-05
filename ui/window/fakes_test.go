package window

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/release"
	"github.com/oernster/ribbonkit/domain/identity/identitytest"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// errPlanted is the failure a stand-in answers when a test asks it to fail.
var errPlanted = errors.New("planted failure")

// scriptedService stands in for the application's service. It answers what a test sets and records
// the calls the window made, so a test reads what the window decided with each answer.
type scriptedService struct {
	choices     ribbon.Choices
	pullOut     bool
	menu        []menus.Item
	settingsFor []menus.Item
	// stuck are the edges EdgeMoves answers would not move the ribbon.
	stuck       map[placement.Edge]bool
	arrangement arranger.Arrangement
	// lastEdge, when set, is what ToLastEdge answers.
	lastEdge *arranger.Arrangement
	// changeErr answers every change; arrangeErr and movedErr answer the arranging calls.
	changeErr  error
	arrangeErr error
	movedErr   error
	// panicOnRearrange makes Rearrange panic, for the window's recover.
	panicOnRearrange bool
	// update answers every update check; panicOnUpdate makes one panic instead. manualChecks records
	// whether each check was asked for; skipped the version each SkipUpdate kept.
	update        release.Status
	panicOnUpdate bool
	manualChecks  []bool
	skipped       []string
	// checked, when set, hears each check as it is made, for a check made on a goroutine.
	checked chan bool

	// previewed and kept are the scales PreviewScale and SetScale were handed, in order.
	previewed []float64
	kept      []int

	calls   []string
	at      []placement.Point
	centred placement.Size
	onTop   []bool
	edges   []placement.Edge
}

func (s *scriptedService) record(call string) { s.calls = append(s.calls, call) }

func (s *scriptedService) change(call string) error {
	s.record(call)
	return s.changeErr
}

func (s *scriptedService) Choices() ribbon.Choices { return s.choices }

func (s *scriptedService) PullOut() bool { return s.pullOut }

func (s *scriptedService) SetPullOut(open bool) error {
	err := s.change("SetPullOut")
	if err == nil {
		s.pullOut = open
	}
	return err
}

func (s *scriptedService) SetColour(ribbon.Colour) error { return s.change("SetColour") }

// SetOrientation takes the choice unless it is refused as one the setting does not offer, as the
// service does: a save that fails still leaves the choice in effect.
func (s *scriptedService) SetOrientation(orientation ribbon.Orientation) error {
	if !errors.Is(s.changeErr, ribbon.ErrUnknownChoice) {
		s.choices.Orientation = orientation
	}
	return s.change("SetOrientation")
}

func (s *scriptedService) SetTheme(ribbon.Theme) error { return s.change("SetTheme") }

func (s *scriptedService) SetAlwaysOnTop(on bool) error {
	s.onTop = append(s.onTop, on)
	s.choices.AlwaysOnTop = on
	return s.change("SetAlwaysOnTop")
}

func (s *scriptedService) SetPinned(on bool) error {
	err := s.change("SetPinned")
	if err == nil {
		s.choices.Pinned = on
	}
	return err
}

func (s *scriptedService) StartWithWindows() (bool, error) { return true, s.changeErr }

func (s *scriptedService) SetStartWithWindows(bool) error { return s.change("SetStartWithWindows") }

func (s *scriptedService) SetScrollbar(int) error { return s.change("SetScrollbar") }

func (s *scriptedService) SetOpacity(percent int) error {
	err := s.change("SetOpacity")
	if err == nil {
		s.choices.Opacity = percent
	}
	return err
}

func (s *scriptedService) PreviewScale(percent float64) error {
	s.previewed = append(s.previewed, percent)
	return s.change("PreviewScale")
}

func (s *scriptedService) SetScale(percent int) error {
	s.kept = append(s.kept, percent)
	return s.change("SetScale")
}

func (s *scriptedService) SetPixelsPerDIP(float64) error { return s.change("SetPixelsPerDIP") }

func (s *scriptedService) ContextMenu() []menus.Item { return s.menu }

func (s *scriptedService) SettingsChoices() []menus.Item { return s.settingsFor }

func (s *scriptedService) CloseRequested() menus.Action { return menus.Hide }

func (s *scriptedService) Launch() (arranger.Arrangement, error) {
	s.record("Launch")
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) Rearrange(at placement.Point) (arranger.Arrangement, error) {
	if s.panicOnRearrange {
		panic("planted panic")
	}
	s.record("Rearrange")
	s.at = append(s.at, at)
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) Moved(at placement.Point) (arranger.Arrangement, error) {
	s.record("Moved")
	s.at = append(s.at, at)
	return s.arrangement, s.movedErr
}

func (s *scriptedService) ToEdge(at placement.Point, edge placement.Edge) (arranger.Arrangement, error) {
	s.record("ToEdge")
	s.at = append(s.at, at)
	s.edges = append(s.edges, edge)
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) EdgeMoves(edge placement.Edge) bool { return !s.stuck[edge] }

// ToLastEdge answers the arrangement against the last edge: lastEdge where a test set one, else the
// scripted arrangement.
func (s *scriptedService) ToLastEdge(at placement.Point) (arranger.Arrangement, error) {
	s.record("ToLastEdge")
	s.at = append(s.at, at)
	if s.lastEdge != nil {
		return *s.lastEdge, s.arrangeErr
	}
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) Centred(at placement.Point, size placement.Size) (arranger.Arrangement, error) {
	s.record("Centred")
	s.at = append(s.at, at)
	s.centred = size
	return s.arrangement, s.arrangeErr
}

// Collapsed answers the tab as a band 8 wide against the arrangement's right side (FR-614).
func (s *scriptedService) Collapsed(full arranger.Arrangement) (arranger.Arrangement, error) {
	s.record("Collapsed")
	return arranger.Arrangement{
		At:   placement.Point{X: full.At.X + full.Size.Width - placement.TabThickness, Y: full.At.Y},
		Size: placement.Size{Width: placement.TabThickness, Height: full.Size.Height},
	}, s.arrangeErr
}

func (s *scriptedService) CheckForUpdate(_ context.Context, manual bool) release.Status {
	if s.panicOnUpdate {
		panic("planted panic")
	}
	if s.checked != nil {
		s.checked <- manual
		return s.update
	}
	s.manualChecks = append(s.manualChecks, manual)
	return s.update
}

func (s *scriptedService) SkipUpdate(version string) error {
	s.skipped = append(s.skipped, version)
	return s.change("SkipUpdate")
}

// Where the tests' ribbon stands, the arrangement the stand-in service answers and the panel's size.
var (
	testRibbonAt = placement.Point{X: 40, Y: 60}
	// testArrange stands flush against its right edge, so an unpinned ribbon arranged there is unpinned
	// in effect (FR-619); testAway is the same ribbon standing against no edge.
	testArrange = arranger.Arrangement{
		At: placement.Point{X: 10, Y: 20}, Size: placement.Size{Width: 300, Height: 90}, Scrolls: true,
		Edge: placement.Right,
	}
	testAway = arranger.Arrangement{At: placement.Point{X: 400, Y: 300}, Size: testArrange.Size}
	// testPanel is every panel's size but Settings', which is testSettingsPanel.
	testPanel         = placement.Size{Width: 560, Height: 760}
	testSettingsPanel = placement.Size{Width: 900, Height: 760}
	// testProduct is the application the tests' window belongs to.
	testProduct = Product{
		App: identitytest.Sample, WindowClass: "SampleRibbonWindow", Version: "1.2.3",
		Author: "A. Author", Copyright: "(c) A. Author", Credits: []Credit{{Name: "Wails", Licence: "MIT", Role: "window"}},
		Licence: "The sample terms.", DonateURL: "https://example.invalid/donate",
	}
)

// testRibbon is the ribbon's window handle once startup has found it: any value that is not none.
const testRibbon = 1

// testUnscaled is the toolkit scale of a desktop that does not scale windows itself.
const testUnscaled = 1

// testThreshold is the stand-in desktop's drag distance.
var testThreshold = placement.Size{Width: 4, Height: 4}

// newTestApp answers a window over a scripted service, started and with its ribbon found, whose
// Wails and desktop calls land in the stand-in answered with it. Menu actions the kit does not know
// land in the stand-in's acted.
func newTestApp(t *testing.T) (*Window, *scriptedService, *window, *bytes.Buffer) {
	t.Helper()
	service := &scriptedService{arrangement: testArrange}
	seen := &window{ribbonAt: testRibbonAt}
	log := &bytes.Buffer{}
	app, _ := New(Config{
		Service: service, Log: log, Product: testProduct,
		Panels: PanelSizes{Settings: testSettingsPanel, Other: testPanel},
		Act:    func(action menus.Action) { seen.acted = append(seen.acted, action) },
	})
	background := context.Background()
	app.ctx.Store(&background)
	app.ribbon = testRibbon
	app.cursor = func() (placement.Point, bool) { return placement.Point{}, false }
	app.emit = func(event string, data ...any) { seen.events = append(seen.events, emitted{event, data}) }
	app.showWindow = func() { seen.shown++ }
	app.hideWindow = func() { seen.hidden++ }
	app.quit = func() { seen.quits++ }
	app.setOnTop = func(on bool) { seen.onTop = append(seen.onTop, on) }
	app.browse = func(address string) error {
		seen.browsed = append(seen.browsed, address)
		return seen.browseErr
	}
	app.showMenu = func(items []menus.Item) { seen.menus = append(seen.menus, items) }
	app.position = func() (placement.Point, error) {
		seen.positions++
		return seen.ribbonAt, seen.readErr
	}
	app.place = func(at placement.Point, size placement.Size) error {
		seen.placed = append(seen.placed, arranger.Arrangement{At: at, Size: size})
		return seen.placeErr
	}
	app.shape = func(parts []placement.Rect) error {
		seen.shapes = append(seen.shapes, parts)
		return nil
	}
	seen.now = testNow
	app.now = func() time.Time { return seen.now }
	app.after = func(wait time.Duration, do func()) func() bool {
		if wait == drawWait {
			seen.drawPending = do
			return func() bool { seen.drawPending = nil; return true }
		}
		// sizeWait and hover.Away are the same length, so the length alone cannot tell the launch's
		// fallback from a collapse; the fallback is armed only while a ready page is not yet shown.
		if wait == sizeWait && app.launch.ready.Load() && !app.launch.shown.Load() {
			seen.sizePending = do
			return func() bool { seen.sizePending = nil; return true }
		}
		seen.pending, seen.waited = do, wait
		return func() bool { seen.pending = nil; return true }
	}
	app.background = func(red, green, blue, alpha uint8) {
		seen.backgrounds = append(seen.backgrounds, [4]uint8{red, green, blue, alpha})
	}
	app.tabFrame = func(tab bool) error {
		seen.tabFrames = append(seen.tabFrames, tab)
		return nil
	}
	app.watchPointer = func(on bool) { seen.watching = append(seen.watching, on) }
	seen.toolkitScale = testUnscaled
	app.toolkitScale = func() int { return seen.toolkitScale }
	// The page's ratio and the drag threshold are the desktop's readings, which differ by platform;
	// the stand-in answers what a test sets, so no test depends on the platform it runs on.
	app.perDIPOf = seen.pixelsPerDIP
	seen.threshold = testThreshold
	app.dragThreshold = func() placement.Size { return seen.threshold }
	service.choices.Pinned = true
	return app, service, seen, log
}

// testNow is the tests' present moment.
var testNow = time.Date(2026, 9, 28, 21, 0, 0, 0, time.UTC)
