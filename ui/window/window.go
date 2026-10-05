// Package window is a ribbon's window as its page and its desktop see it: placing, collapsing to the
// tab and opening, the corner grip, the opacity, the panels, the pull out beside the ribbon, the menus,
// the update check and the window's own life. An application embeds the Window in the object it hands
// Wails, so every exported method of Window is page API; the application's own methods reach the
// window through the Control, which is never embedded and so never bound.
//
// FR numbers are TimeRibbon's REQUIREMENTS.md, where each rule was first specified.
package window

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/release"
	"github.com/oernster/ribbonkit/application/shell"
	"github.com/oernster/ribbonkit/domain/identity"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// Service is what the window asks of the application layer: the ribbon's choices, the arranging of
// its window, its menus and the update check. An application's service is one; a test's scripted
// stand-in is another, which reads what the window decided with each answer.
type Service interface {
	// Choices answers the ribbon's choices as they stand.
	Choices() ribbon.Choices
	// PullOut answers whether the pull out is open beside the ribbon (FR-903).
	PullOut() bool
	SetPullOut(open bool) error
	SetColour(colour ribbon.Colour) error
	SetOrientation(orientation ribbon.Orientation) error
	SetTheme(theme ribbon.Theme) error
	SetAlwaysOnTop(on bool) error
	SetPinned(on bool) error
	StartWithWindows() (bool, error)
	SetStartWithWindows(on bool) error
	SetScrollbar(dip int) error
	SetOpacity(percent int) error
	PreviewScale(percent float64) error
	SetScale(percent int) error
	SetPixelsPerDIP(scale float64) error

	ContextMenu() []menus.Item
	SettingsChoices() []menus.Item
	CloseRequested() menus.Action

	Launch() (arranger.Arrangement, error)
	Rearrange(at placement.Point) (arranger.Arrangement, error)
	Moved(at placement.Point) (arranger.Arrangement, error)
	ToEdge(at placement.Point, edge placement.Edge) (arranger.Arrangement, error)
	// EdgeMoves answers whether ToEdge would move the ribbon from where it stands (FR-408, FR-412).
	EdgeMoves(edge placement.Edge) bool
	ToLastEdge(at placement.Point) (arranger.Arrangement, error)
	Centred(at placement.Point, size placement.Size) (arranger.Arrangement, error)
	Collapsed(full arranger.Arrangement) (arranger.Arrangement, error)

	CheckForUpdate(ctx context.Context, manual bool) release.Status
	SkipUpdate(version string) error
}

// Product is what the window says about the application it belongs to: its names, the window class
// the desktop finds the ribbon by and what Help shows (FR-607, FR-608).
type Product struct {
	App         identity.App
	WindowClass string
	Version     string
	Author      string
	Copyright   string
	Credits     []Credit
	// Licence is the whole of the terms the application is released under.
	Licence string
	// DonateURL is the donation page the button hands to the browser.
	DonateURL string
}

// Credit is one component the application ships.
type Credit struct {
	Name    string
	Licence string
	Role    string
}

// Config is everything the window is built from.
type Config struct {
	Service Service
	Desktop shell.Desktop
	Log     io.Writer
	Panels  PanelSizes
	Product Product
	// Act carries out a menu action of the application's own, one the kit does not know; nil when
	// the application adds none.
	Act func(menus.Action)
}

// Window is the ribbon's window. Embedded in the object Wails binds, its exported methods are what
// the page calls.
type Window struct {
	service Service
	desktop shell.Desktop
	log     io.Writer
	panels  PanelSizes
	product Product
	actOn   func(menus.Action)

	// The window's calls into Wails and the desktop. Each is a field so a test can stand in for it
	// and read what the window did; New points them at the real calls in wails_calls.go and at the
	// desktop port.
	emit       func(event string, data ...any)
	showWindow func()
	hideWindow func()
	quit       func()
	setOnTop   func(on bool)
	browse     func(address string) error
	showMenu   func(items []menus.Item)
	position   func() (placement.Point, error)
	place      func(at placement.Point, size placement.Size) error
	shape      func(parts []placement.Rect) error
	background func(red, green, blue, alpha uint8)
	paint      paintState
	// The unpinned ribbon's calls (FR-613 to FR-618): the time, a timer that answers its own stop,
	// the tab's frame and the desktop's reporting of the pointer.
	now          func() time.Time
	after        func(wait time.Duration, do func()) func() bool
	tabFrame     func(tab bool) error
	watchPointer func(on bool)
	// toolkitScale answers the toolkit's own window scale, which the page's ratio is divided by;
	// perDIPOf turns the page's ratio into window pixels with it. dragThreshold is how far the pointer
	// moves before a press becomes a drag.
	toolkitScale  func() int
	perDIPOf      func(pageRatio float64, toolkitScale int) float64
	dragThreshold func() placement.Size
	// cursor answers the desktop's own reading of the pointer, which the grip's drag prefers.
	cursor func() (placement.Point, bool)
	grip   gripDrag
	// launch is what the launched ribbon's first showing waits for (launch_show.go); lastPlaced is
	// where the window was last put, to tell when the desktop showed it somewhere else.
	launch     launchShow
	lastPlaced placedWindow

	ctx       context.Context
	ribbon    shell.Window
	trayUp    atomic.Bool
	visible   atomic.Bool
	quitting  atomic.Bool
	panelOpen atomic.Bool
	scrolls   atomic.Bool
	// panelWidth is the open panel's width in DIP, which fitting its height keeps.
	panelWidth atomic.Int64
	// pixelsPerDIP is the window pixels to each of the page's units that the service sizes windows
	// with, as math.Float64bits; zero until the page has reported its ratio (SetPixelRatio).
	pixelsPerDIP atomic.Uint64

	// updates holds the update check's timing and the outcome it last offered (FR-509).
	updates updateWatch

	// unpin is the unpinned ribbon's hover state and what the window shows of it (FR-613 to FR-618).
	unpin unpinned
}

// New answers the window built from config with the Control over it. Every call into the desktop goes
// through a closure rather than a method value, so a window built with no desktop, as the tests build
// it, fails only if one is actually made.
func New(config Config) (*Window, *Control) {
	desk := config.Desktop
	built := &Window{
		service: config.Service, desktop: desk, log: config.Log, panels: config.Panels,
		product: config.Product, actOn: config.Act,
		updates: updateWatch{delay: updateCheckDelay, every: updateCheckEvery},
	}
	if built.actOn == nil {
		built.actOn = func(menus.Action) {}
	}
	built.emit = built.emitToWails
	built.showWindow = built.showInWails
	built.hideWindow = built.hideInWails
	built.quit = built.quitWails
	built.setOnTop = built.setOnTopInWails
	built.browse = func(address string) error { return desk.OpenInBrowser(address) }
	built.showMenu = func(items []menus.Item) { desk.ShowMenu(items) }
	built.position = func() (placement.Point, error) { return desk.Position(built.ribbon) }
	built.place = func(at placement.Point, size placement.Size) error { return desk.Place(built.ribbon, at, size) }
	built.shape = func(parts []placement.Rect) error { return desk.Shape(built.ribbon, parts) }
	built.background = built.backgroundInWails
	built.now = time.Now
	built.after = func(wait time.Duration, do func()) func() bool { return time.AfterFunc(wait, do).Stop }
	built.tabFrame = func(tab bool) error { return desk.SetTabFrame(built.ribbon, tab) }
	built.watchPointer = func(on bool) { desk.TrackPointer(built.ribbon, on) }
	built.toolkitScale = func() int { return desk.ToolkitScale() }
	built.perDIPOf = func(pageRatio float64, toolkitScale int) float64 {
		return desk.PixelsPerDIP(pageRatio, toolkitScale)
	}
	built.dragThreshold = func() placement.Size { return desk.DragThreshold() }
	built.cursor = func() (placement.Point, bool) { return desk.Cursor() }
	// The window opens as the full ribbon; it is collapsed only once it has been arranged.
	built.unpin.shownOpen = true
	return built, &Control{window: built}
}
