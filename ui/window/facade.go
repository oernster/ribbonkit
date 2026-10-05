package window

// The page's calls about the ribbon itself. Every method runs one use case, then does what the window
// needs afterwards. Methods that can be refused answer an error, which rejects the page's promise.

import (
	"fmt"
	"math"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// Events the page listens for.
const (
	eventRefresh   = "refresh"
	eventOpenPanel = "open-panel"
)

// Which panel an open-panel event asks for: Settings, About, Licence or an update check's outcome,
// which travels with it (CON-6, FR-508, FR-509). An application may open panels of its own the same
// way (Control.ShowPanel).
const (
	openAtSettings = "settings"
	openAtAbout    = "about"
	openAtLicence  = "licence"
	openAtUpdate   = "update"
)

// SetColour chooses the colour scheme (FR-611).
func (a *Window) SetColour(colour string) error {
	return a.refitted(a.service.SetColour(ribbon.Colour(colour)))
}

// SetOrientation chooses horizontal or vertical (FR-103), then puts the ribbon against that
// orientation's home edge (FR-409). A choice that did not take, as one the setting does not offer,
// fits the ribbon where it stands. One whose save failed has still taken, so it moves.
func (a *Window) SetOrientation(orientation string) error {
	chosen := ribbon.Orientation(orientation)
	err := a.service.SetOrientation(chosen)
	edge, known := ribbon.HomeEdge(chosen)
	if !known || a.service.Choices().Orientation != chosen {
		a.contentChanged()
		return err
	}
	a.toEdge(edge)
	return err
}

// SetTheme chooses system, light or dark (FR-606).
func (a *Window) SetTheme(theme string) error {
	return a.refitted(a.service.SetTheme(ribbon.Theme(theme)))
}

// SetAlwaysOnTop turns Always on Top on or off and applies it at once (FR-505).
func (a *Window) SetAlwaysOnTop(on bool) error {
	err := a.service.SetAlwaysOnTop(on)
	a.applyAlwaysOnTop()
	return a.refitted(err)
}

// StartWithWindows answers whether the Start with Windows value is present (FR-605).
func (a *Window) StartWithWindows() (bool, error) { return a.service.StartWithWindows() }

// SetStartWithWindows writes or removes the Start with Windows value (FR-605).
func (a *Window) SetStartWithWindows(on bool) error { return a.service.SetStartWithWindows(on) }

// SetScrollbar takes the thickness in DIP of the scroll bar the page draws, which it measures once
// it has loaded, then fits the ribbon with room for it (FR-106).
func (a *Window) SetScrollbar(dip int) error { return a.refitted(a.service.SetScrollbar(dip)) }

// SetPixelRatio takes the page's devicePixelRatio, which it reports once it has loaded and again
// whenever it changes, then fits the window to the page as it is really drawn. Windows' text size
// enlarges the page without changing the display's DPI, so the DPI alone left the page cut off. The
// refit can move the pull out to the ribbon's other side, as when a drag carries the window onto a
// display at other scaling, so the page draws again (measured 2026-10-05).
func (a *Window) SetPixelRatio(ratio float64) error {
	perDIP := a.perDIPOf(ratio, a.toolkitScale())
	err := a.service.SetPixelsPerDIP(perDIP)
	if err == nil {
		a.pixelsPerDIP.Store(math.Float64bits(perDIP))
	}
	err = a.redrawn(a.refitted(err))
	if err == nil {
		a.pageScaled()
	}
	return err
}

// SetBackground takes the colour the page paints behind everything, which it reports once it has
// loaded and again whenever the scheme or theme changes it, so the window shows that colour rather
// than white while the page catches up with a new size (measured 2026-09-29). The colours live in the
// page's CSS alone; Go only passes this one on. A channel outside a byte is refused.
func (a *Window) SetBackground(red, green, blue int) error {
	for _, channel := range []int{red, green, blue} {
		if channel < 0 || channel > math.MaxUint8 {
			return fmt.Errorf("the page's background rgb(%d, %d, %d) is not a colour", red, green, blue)
		}
	}
	a.paint.guard.Lock()
	a.paint.colour, a.paint.known = [3]uint8{uint8(red), uint8(green), uint8(blue)}, true
	a.paint.guard.Unlock()
	a.repaint()
	return nil
}

// ShowContextMenu shows the ribbon's right-click menu as a native menu at the cursor (FR-108). While
// it is open the ribbon does not collapse (FR-616); the desktop reports it closed.
func (a *Window) ShowContextMenu() {
	a.menuShown()
	a.showMenu(a.service.ContextMenu())
}

// OpenDonation hands the donation page to the desktop's browser. The application never fetches it,
// so the button adds no network request to the update check's one (NFR-S-1). Where the desktop
// cannot open it, the refusal says why and gives the address, so it can still be reached by hand.
// The words name no system, since every platform's desktop can refuse.
func (a *Window) OpenDonation() error {
	address := a.product.DonateURL
	if err := a.browse(address); err != nil {
		return fmt.Errorf("your browser could not be opened on the donation page (%w). There may be no default browser set; the page is %s", err, address)
	}
	return nil
}

// About answers what the About panel shows (FR-607).
func (a *Window) About() aboutDTO {
	credits := make([]creditDTO, 0, len(a.product.Credits))
	for _, credit := range a.product.Credits {
		credits = append(credits, creditDTO(credit))
	}
	return aboutDTO{
		Name: a.product.App.Name, Version: a.product.Version, Author: a.product.Author,
		Copyright: a.product.Copyright, Credits: credits,
	}
}

// Licence answers the whole of the terms the application is released under (FR-608).
func (a *Window) Licence() string { return a.product.Licence }

// Hide hides the ribbon (FR-504).
func (a *Window) Hide() { a.hide() }

// refitted fits the ribbon after a change, then answers err. A change whose save failed raises a
// notice, which is one more cell to fit (FR-707); one that saved may have ended an earlier notice.
func (a *Window) refitted(err error) error {
	a.contentChanged()
	return err
}

// contentChanged fits the ribbon to what it now holds where it stands; where its length changed it
// is centred along that length again (FR-104, FR-105). While a panel is open the window is that
// panel, so the ribbon is fitted when it closes instead.
func (a *Window) contentChanged() {
	if a.panelOpen.Load() || a.ribbon == 0 {
		return
	}
	a.rearrange()
}
