package arranger

import (
	"errors"
	"sync"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// errPlanted is the failure a fake answers when a test asks it to fail.
var errPlanted = errors.New("planted failure")

// The cells the tests arrange, in DIP: a digital clock's, a taller analogue one's and the small
// size's, as TimeRibbon draws them, so the arithmetic in each test is worked in familiar numbers.
var (
	digitalCell  = placement.Size{Width: 160, Height: 90}
	analogueCell = placement.Size{Width: 160, Height: 150}
	smallCell    = placement.Size{Width: 120, Height: 60}
)

// The padding round the cells and the depth of a pull out handle's lane, in DIP.
const (
	testPadding = 8
	testLane    = 12
)

// Two monitors side by side: a primary at 100 percent and a secondary at 150 percent to its right.
var (
	primaryMonitor = placement.Monitor{
		Device: `\\.\DISPLAY1`, Work: placement.Rect{Right: 1920, Bottom: 1032},
		DPI: placement.BaseDPI, Primary: true,
	}
	secondaryMonitor = placement.Monitor{
		Device: `\\.\DISPLAY2`, Work: placement.Rect{Left: 1920, Right: 4480, Bottom: 1392}, DPI: 144,
	}
)

type fakeMonitors struct {
	monitors []placement.Monitor
	err      error
}

func (f *fakeMonitors) Monitors() ([]placement.Monitor, error) { return f.monitors, f.err }

// fakeHost is an application holding the ribbon's choices and content. Like TimeRibbon, it keeps a
// change in effect when its save fails and raises a notice drawn as one more cell until a save
// succeeds or the notice is dismissed (FR-707).
type fakeHost struct {
	mutex   sync.Mutex
	choices ribbon.Choices
	content Content
	saveErr error
	notice  bool
	saved   []ribbon.Choices
}

func (h *fakeHost) Ribbon() (ribbon.Choices, Content) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	content := h.content
	if h.notice {
		content.Cells++
	}
	return h.choices, content
}

func (h *fakeHost) ChangeRibbon(edit func(ribbon.Choices) ribbon.Choices) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.choices = edit(h.choices)
	if h.saveErr != nil {
		h.notice = true
		return h.saveErr
	}
	h.notice = false
	h.saved = append(h.saved, h.choices)
	return nil
}

// edit changes the content as the application would, such as by adding a clock.
func (h *fakeHost) edit(change func(*Content)) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	change(&h.content)
}

// failSaves makes every save fail with err; a nil err lets saves succeed again.
func (h *fakeHost) failSaves(err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.saveErr = err
}

// dismiss clears the notice, as reading it does.
func (h *fakeHost) dismiss() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.notice = false
}

// current answers the choices in effect.
func (h *fakeHost) current() ribbon.Choices {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.choices
}

// last answers the choices saved most recently, failing the test when none were.
func (h *fakeHost) last(t *testing.T) ribbon.Choices {
	t.Helper()
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if len(h.saved) == 0 {
		t.Fatal("nothing was saved")
	}
	return h.saved[len(h.saved)-1]
}

// saves counts the saves that succeeded.
func (h *fakeHost) saves() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return len(h.saved)
}

// fakeNeighbours stands in for the other ribbons running: the rectangles they hold and every
// footprint this ribbon held (FR-412).
type fakeNeighbours struct {
	mutex sync.Mutex
	taken []placement.Rect
	held  [][]placement.Rect
}

func (n *fakeNeighbours) Taken() []placement.Rect {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.taken
}

func (n *fakeNeighbours) Hold(occupied []placement.Rect) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.held = append(n.held, occupied)
}

// occupy sets the rectangles the other ribbons hold.
func (n *fakeNeighbours) occupy(taken ...placement.Rect) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.taken = taken
}

// lastHeld answers the footprint this ribbon held most recently, failing the test when none was.
func (n *fakeNeighbours) lastHeld(t *testing.T) []placement.Rect {
	t.Helper()
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if len(n.held) == 0 {
		t.Fatal("nothing was held")
	}
	return n.held[len(n.held)-1]
}

// rig is an arranger over fakes, with the fakes kept to inspect.
type rig struct {
	arranger   *Arranger
	host       *fakeHost
	monitors   *fakeMonitors
	neighbours *fakeNeighbours
}

// newRig answers an arranger over both monitors of a ribbon holding choices and content, alone until
// its neighbours are given rectangles.
func newRig(choices ribbon.Choices, content Content) rig {
	host := &fakeHost{choices: choices, content: content}
	monitors := &fakeMonitors{monitors: []placement.Monitor{primaryMonitor, secondaryMonitor}}
	neighbours := &fakeNeighbours{}
	return rig{arranger: New(host, monitors, neighbours), host: host, monitors: monitors, neighbours: neighbours}
}

// horizontal answers the first-run choices running horizontally; the arithmetic in the tests is
// worked for horizontal cells unless they say otherwise.
func horizontal() ribbon.Choices {
	choices := ribbon.Defaults()
	choices.Orientation = ribbon.Horizontal
	return choices
}

// vertical answers the first-run choices, which run vertically (FR-103).
func vertical() ribbon.Choices {
	return ribbon.Defaults()
}

// cells answers n digital cells.
func cells(n int) Content {
	return Content{Cell: digitalCell, Cells: n, Padding: testPadding}
}

// draggedTo answers choices in orientation stored where a drag left the ribbon at at on the primary.
func draggedTo(orientation ribbon.Orientation, at placement.Point) ribbon.Choices {
	choices := ribbon.Defaults()
	choices.Orientation = orientation
	choices.Placement = &placement.Stored{Device: primaryMonitor.Device, DPI: placement.BaseDPI, Offset: at}
	return choices
}
