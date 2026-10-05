package window

import (
	"math"
	"sync"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// gripDrag is a drag of the corner grip under way (FR-623): where the pointer began, in the page's
// units; the scale and the ribbon's thickness then; the scale last previewed, zero before the
// first. guard serialises the drag's calls, which Wails runs on goroutines of their own, so the
// scale kept at the end can never be overtaken by a preview still on its way.
type gripDrag struct {
	guard     sync.Mutex
	active    bool
	fromX     float64
	fromY     float64
	began     float64
	thickness float64
	reached   float64
}

// BeginScale starts a drag of the grip: thickness is the ribbon's across its orientation and x, y
// where the page read the pointer, both in the page's units.
func (a *Window) BeginScale(thickness, x, y float64) {
	a.grip.guard.Lock()
	defer a.grip.guard.Unlock()
	a.grip.active, a.grip.thickness, a.grip.reached = true, thickness, 0
	a.grip.began = float64(a.service.Choices().Scale)
	a.grip.fromX, a.grip.fromY = a.pointerAt(x, y)
}

// DragScale previews the scale the grip has reached, where it differs from the last one previewed.
func (a *Window) DragScale(x, y float64) error {
	a.grip.guard.Lock()
	defer a.grip.guard.Unlock()
	if !a.grip.active {
		return nil
	}
	scale := a.gripScale(x, y)
	if scale == a.grip.reached || (a.grip.reached == 0 && scale == a.grip.began) {
		return nil
	}
	a.grip.reached = scale
	return a.PreviewScale(scale)
}

// EndScale keeps the scale the grip was let go at, to the nearest whole percent; a press that
// previewed nothing keeps nothing.
func (a *Window) EndScale(x, y float64) error {
	a.grip.guard.Lock()
	defer a.grip.guard.Unlock()
	if !a.grip.active {
		return nil
	}
	a.grip.active = false
	if a.grip.reached == 0 {
		return nil
	}
	return a.SetScale(int(math.Round(a.gripScale(x, y))))
}

// gripScale answers the scale the pointer at x, y has drawn the ribbon at, moving its far side.
func (a *Window) gripScale(x, y float64) float64 {
	atX, atY := a.pointerAt(x, y)
	moved := atY - a.grip.fromY
	if a.service.Choices().Orientation == ribbon.Vertical {
		moved = atX - a.grip.fromX
	}
	return ribbon.ScaleAfter(a.grip.began, a.grip.thickness, moved)
}

// pointerAt answers where the pointer is in the page's units: the desktop's own reading wherever it
// can give one and the page has reported its ratio, else x, y as the page read them.
func (a *Window) pointerAt(x, y float64) (float64, float64) {
	perDIP := math.Float64frombits(a.pixelsPerDIP.Load())
	if at, ok := a.cursor(); ok && perDIP > 0 {
		return float64(at.X) / perDIP, float64(at.Y) / perDIP
	}
	return x, y
}

// PreviewScale draws the ribbon at percent while its grip is dragged, keeping nothing, then fits
// the window to it and tells the page to draw again (FR-623).
func (a *Window) PreviewScale(percent float64) error {
	return a.redrawn(a.refitted(a.service.PreviewScale(percent)))
}

// SetScale keeps the scale the grip's drag ended at, then fits the window to it and tells the page
// to draw again (FR-623).
func (a *Window) SetScale(percent int) error {
	return a.redrawn(a.refitted(a.service.SetScale(percent)))
}
