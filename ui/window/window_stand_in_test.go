package window

// The window stand-in: what the facade asked of Wails and the desktop, kept apart from the scripted
// service it is used beside.

import (
	"time"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
)

// emitted is one event the facade sent the page.
type emitted struct {
	event string
	data  []any
}

// window records what the facade asked of Wails and the desktop, standing in for both.
type window struct {
	events    []emitted
	shown     int
	hidden    int
	quits     int
	onTop     []bool
	browsed   []string
	menus     [][]menus.Item
	placed    []arranger.Arrangement
	shapes    [][]placement.Rect
	ribbonAt  placement.Point
	readErr   error
	placeErr  error
	browseErr error
	positions int
	// The unpinned ribbon's calls: the frames asked for, the pointer watching asked for, the time
	// the tests set and the timers pending, which a test fires by hand: hover's, then the fallback
	// that grows an opening ribbon whose page has not said it has drawn.
	tabFrames   []bool
	watching    []bool
	backgrounds [][4]uint8
	now         time.Time
	pending     func()
	waited      time.Duration
	drawPending func()
	// toolkitScale is the toolkit's window scale the desktop reports; unscaled unless a test sets it.
	toolkitScale int
	// perDIP, when set, is what the desktop answers for the window pixels to each of the page's units;
	// otherwise it answers the page's ratio as it is. perDIPAsked is each ratio and scale it was
	// asked about. threshold is the desktop's drag distance.
	perDIP      float64
	perDIPAsked []perDIPAsk
	threshold   placement.Size
	// sizePending is the launch's fallback for a page that never sizes the ribbon.
	sizePending func()
	// acted is each menu action the window handed the application as one of its own.
	acted []menus.Action
}

// perDIPAsk is one question put to the desktop about the page's ratio.
type perDIPAsk struct {
	ratio float64
	scale int
}

// pixelsPerDIP answers as the desktop does for the page's ratio at the toolkit's scale, recording it.
func (w *window) pixelsPerDIP(ratio float64, scale int) float64 {
	w.perDIPAsked = append(w.perDIPAsked, perDIPAsk{ratio, scale})
	if w.perDIP != 0 {
		return w.perDIP
	}
	return ratio
}

// sawEvent reports whether the facade sent event with data first, when data is given.
func (w *window) sawEvent(event string, data ...any) bool {
	for _, sent := range w.events {
		if sent.event == event && (len(data) == 0 || (len(sent.data) > 0 && sent.data[0] == data[0])) {
			return true
		}
	}
	return false
}
