package window

// The window as a panel: Settings, About, Licence, an update check's outcome or one of the
// application's own (CON-6).

import (
	"fmt"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// PanelSizes are the sizes in DIP the window opens at as a panel: Settings wide enough to lay its
// choices side by side (FR-625), every other panel of the kit's narrower, where its text reads better.
type PanelSizes struct {
	Settings placement.Size
	Other    placement.Size
	// Own are the application's own panels, each by the word the page names it by, such as a city's
	// detail; nil where it has none. A word the kit names a panel by stays the kit's.
	Own map[string]placement.Size
}

// of answers the size panel opens at, panel being the word the page names it by; a word naming no
// panel is refused.
func (p PanelSizes) of(panel string) (placement.Size, error) {
	switch panel {
	case openAtSettings:
		return p.Settings, nil
	case openAtAbout, openAtLicence, openAtUpdate:
		return p.Other, nil
	}
	if size, ok := p.Own[panel]; ok {
		return size, nil
	}
	return placement.Size{}, fmt.Errorf("%w: a panel named %q", ribbon.ErrUnknownChoice, panel)
}

// OpenPanel turns the window into panel, centred on the ribbon's display (CON-6). A panel holds an
// unpinned ribbon open until it closes (FR-616).
func (a *Window) OpenPanel(panel string) error {
	size, err := a.panels.of(panel)
	if err != nil {
		return err
	}
	a.panelWidth.Store(int64(size.Width))
	if !a.panelOpen.Swap(true) {
		a.hold(true)
	}
	at, err := a.ribbonAt()
	if err != nil {
		return err
	}
	arranged, err := a.service.Centred(at, size)
	if err != nil {
		return err
	}
	a.report("giving the panel its frame", a.tabFrame(false))
	return a.placeWhole(arranged.At, arranged.Size)
}

// FitPanel makes an open panel as tall as its content in DIP, the page's measure of it, re-centred on
// the display it is on; never taller than that display's work area, where it scrolls instead
// (FR-621). With no panel open there is nothing to fit, nor with no height; a negative one is refused.
func (a *Window) FitPanel(height int) error {
	if height < 0 {
		return fmt.Errorf("%w: a panel %d tall", placement.ErrNegativeLength, height)
	}
	if !a.panelOpen.Load() || height == 0 {
		return nil
	}
	at, err := a.position()
	if err != nil {
		return err
	}
	arranged, err := a.service.Centred(at, placement.Size{Width: int(a.panelWidth.Load()), Height: height})
	if err != nil {
		return err
	}
	return a.placeWhole(arranged.At, arranged.Size)
}

// ClosePanel returns the window to the ribbon, where it was last left (CON-6, FR-405), then lets an
// unpinned one collapse once the pointer is away (FR-616).
func (a *Window) ClosePanel() error {
	wasOpen := a.panelOpen.Swap(false)
	err := a.placeLaunched()
	if wasOpen {
		a.release()
	}
	return err
}
