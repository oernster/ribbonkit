package menus

import (
	"slices"
	"strings"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// The items every ribbon's menus hold, built from the ribbon's choices. Each application lists them
// in its own order among items of its own. The words shown for each live here alone.

// The actions of the Orientation submenu.
const (
	OrientHorizontal Action = "horizontal"
	OrientVertical   Action = "vertical"
)

// colourPrefix begins each Colour item's action, which ends in the scheme it chooses.
const colourPrefix = "colour-"

// Item words, one home each.
const (
	labelShow        = "Show ribbon"
	labelHide        = "Hide ribbon"
	labelSettings    = "Settings"
	labelPosition    = "Position"
	labelLeftEdge    = "Centre on left edge"
	labelRightEdge   = "Centre on right edge"
	labelTopEdge     = "Centre on top edge"
	labelBottomEdge  = "Centre on bottom edge"
	labelAlwaysOnTop = "Always on top"
	labelPin         = "Pin ribbon"
	labelHelp        = "Help"
	labelAbout       = "About"
	labelLicence     = "Licence"
	labelUpdates     = "Check for updates"
	labelExit        = "Exit"
	labelColour      = "Colour"
	labelOrientation = "Orientation"
	labelHorizontal  = "Horizontal"
	labelVertical    = "Vertical"
)

// colourLabels are the Colour items' words, one home each.
var colourLabels = map[ribbon.Colour]string{
	ribbon.Classic: "Classic", ribbon.Neon: "Neon", ribbon.Ocean: "Ocean",
	ribbon.Sunset: "Sunset", ribbon.Forest: "Forest", ribbon.Amber: "Amber",
	ribbon.Ruby: "Ruby", ribbon.Indigo: "Indigo", ribbon.Berry: "Berry",
	ribbon.Contrast: "Contrast",
}

// orientationActions maps each Orientation item to the orientation it chooses.
var orientationActions = map[Action]ribbon.Orientation{
	OrientHorizontal: ribbon.Horizontal, OrientVertical: ribbon.Vertical,
}

// Visibility is the tray's first item for a ribbon that is or is not visible (FR-502): it names the
// opposite of what is, so it says what pressing it will do.
func Visibility(visible bool) Item {
	if visible {
		return HideItem()
	}
	return Item{Action: Show, Label: labelShow}
}

// HideItem hides the ribbon; the application keeps running.
func HideItem() Item { return Item{Action: Hide, Label: labelHide} }

// SettingsItem opens Settings.
func SettingsItem() Item { return Item{Action: Settings, Label: labelSettings} }

// ExitItem ends the application (FR-108, FR-502).
func ExitItem() Item { return Item{Action: Exit, Label: labelExit} }

// HelpItem is the Help submenu (FR-508, FR-509).
func HelpItem() Item {
	return Item{Label: labelHelp, Children: []Item{
		{Action: About, Label: labelAbout},
		{Action: Licence, Label: labelLicence},
		{Action: Updates, Label: labelUpdates},
	}}
}

// PositionItem is the Position submenu (FR-408): the two edges along which the ribbon runs its
// length, so a vertical ribbon is offered the left and right edges and a horizontal one the top and
// bottom.
func PositionItem(orientation ribbon.Orientation) Item {
	children := []Item{{Action: TopEdge, Label: labelTopEdge}, {Action: BottomEdge, Label: labelBottomEdge}}
	if orientation == ribbon.Vertical {
		children = []Item{{Action: LeftEdge, Label: labelLeftEdge}, {Action: RightEdge, Label: labelRightEdge}}
	}
	return Item{Label: labelPosition, Children: children}
}

// AlwaysOnTopItem is ticked while the ribbon stays above other windows (FR-505).
func AlwaysOnTopItem(on bool) Item {
	return Item{Action: AlwaysOnTop, Label: labelAlwaysOnTop, Checkable: true, Checked: on}
}

// PinItem follows Always on top, ticked while the ribbon is pinned (FR-613).
func PinItem(on bool) Item {
	return Item{Action: Pin, Label: labelPin, Checkable: true, Checked: on}
}

// ColourItem is the Colour submenu, the current scheme ticked (FR-611).
func ColourItem(current ribbon.Colour) Item {
	children := make([]Item, 0, len(ribbon.Colours))
	for _, colour := range ribbon.Colours {
		children = append(children, Item{
			Action: Action(colourPrefix + string(colour)), Label: colourLabels[colour],
			Checkable: true, Checked: colour == current,
		})
	}
	return Item{Label: labelColour, Children: children}
}

// ColourOf answers the scheme a Colour item chooses; false for any other action.
func ColourOf(action Action) (ribbon.Colour, bool) {
	name, found := strings.CutPrefix(string(action), colourPrefix)
	colour := ribbon.Colour(name)
	return colour, found && slices.Contains(ribbon.Colours, colour)
}

// OrientationItem is the Orientation submenu, the current orientation ticked.
func OrientationItem(current ribbon.Orientation) Item {
	return Item{Label: labelOrientation, Children: []Item{
		{Action: OrientHorizontal, Label: labelHorizontal, Checkable: true, Checked: current == ribbon.Horizontal},
		{Action: OrientVertical, Label: labelVertical, Checkable: true, Checked: current == ribbon.Vertical},
	}}
}

// OrientationOf answers the orientation an Orientation item chooses; false for any other action.
func OrientationOf(action Action) (ribbon.Orientation, bool) {
	orientation, ok := orientationActions[action]
	return orientation, ok
}
