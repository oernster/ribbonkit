package window

import (
	"math"
	"sync"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// paintState is the page's background as it last reported it, kept so a change of opacity can paint
// the window again at once rather than wait for the page's next report.
type paintState struct {
	guard  sync.Mutex
	colour [3]uint8
	known  bool
}

// SetOpacity chooses how opaque the window is drawn (FR-622), then paints the window's own
// background again for it. The page draws itself at the opacity from the snapshot.
func (a *Window) SetOpacity(percent int) error {
	err := a.service.SetOpacity(percent)
	a.repaint()
	return err
}

// repaint paints the window in the page's last reported background: opaque while the window is
// wholly opaque, so a window catching up with a new size shows the page's colour rather than
// white; clear otherwise, since the window's own paint shows behind the page wherever the page is
// less than opaque and would hide the desktop the chosen opacity lets through.
func (a *Window) repaint() {
	a.paint.guard.Lock()
	defer a.paint.guard.Unlock()
	if !a.paint.known {
		return
	}
	var alpha uint8
	if a.service.Choices().Opacity >= ribbon.MaxOpacity {
		alpha = math.MaxUint8
	}
	a.background(a.paint.colour[0], a.paint.colour[1], a.paint.colour[2], alpha)
}
