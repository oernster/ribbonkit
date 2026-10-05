// Package ribbon holds the choices every ribbon has, whatever its cells show: where it stands and
// runs, how it is coloured and drawn, whether it waits as a tab and which release was skipped. An
// application embeds Choices in its own settings beside what its cells need.
//
// Every operation answers a new value and leaves the receiver as it was. FR numbers are
// TimeRibbon's REQUIREMENTS.md, where each rule was first specified.
package ribbon

import (
	"errors"
	"slices"

	"github.com/oernster/ribbonkit/domain/placement"
)

// ErrUnknownChoice is answered when a choice is given a value it does not offer: one Normalised
// would replace.
var ErrUnknownChoice = errors.New("not one of the values this setting offers")

// Colour is the colour scheme the ribbon is drawn in (FR-611).
type Colour string

// The colour schemes. Classic is the look the ribbon has always had.
const (
	Classic  Colour = "classic"
	Neon     Colour = "neon"
	Ocean    Colour = "ocean"
	Sunset   Colour = "sunset"
	Forest   Colour = "forest"
	Amber    Colour = "amber"
	Ruby     Colour = "ruby"
	Indigo   Colour = "indigo"
	Berry    Colour = "berry"
	Contrast Colour = "contrast"
)

// Colours lists the colour schemes in the order they are offered.
var Colours = []Colour{Classic, Neon, Ocean, Sunset, Forest, Amber, Ruby, Indigo, Berry, Contrast}

// Orientation is the direction cells run in (FR-103).
type Orientation string

// The orientations.
const (
	Horizontal Orientation = "horizontal"
	Vertical   Orientation = "vertical"
)

// homeEdges is each orientation's home edge (FR-409): a horizontal ribbon goes to the top, a
// vertical one to the right.
var homeEdges = map[Orientation]placement.Edge{Horizontal: placement.Top, Vertical: placement.Right}

// HomeEdge answers the edge a ribbon of orientation goes to when that orientation is chosen and
// wherever it has no place of its own (FR-403, FR-409); false for an orientation not offered.
func HomeEdge(orientation Orientation) (placement.Edge, bool) {
	edge, ok := homeEdges[orientation]
	return edge, ok
}

// Theme is light, dark or the system's (FR-606).
type Theme string

// The themes.
const (
	System Theme = "system"
	Light  Theme = "light"
	Dark   Theme = "dark"
)

// The scale the cells may be drawn at on top of their size, in percent (FR-623). At MinScale an
// 11 px text draws at about 8 px, the least that stays readable; at MaxScale TimeRibbon's thickest
// ribbon is 408 DIP across, which a 720 line display still holds. WholeScale draws cells as they are.
const (
	MinScale   = 75
	WholeScale = 100
	MaxScale   = 200
)

// The opacity a ribbon's background may be drawn at, in percent (FR-622): wholly opaque at most; at
// least faint enough to see through while never so faint the ribbon cannot be seen or found again
// (Oliver, 2026-09-29).
const (
	MinOpacity = 20
	MaxOpacity = 100
)

// Choices is every choice about the ribbon itself.
type Choices struct {
	Colour      Colour
	Orientation Orientation
	Theme       Theme
	AlwaysOnTop bool
	// Pinned keeps the ribbon shown in full; unpinned, it waits as a tab (FR-613).
	Pinned bool
	// SkippedUpdate is the release the user chose to skip, which the automatic update check never
	// offers again (FR-509); empty when none has been skipped.
	SkippedUpdate string
	// Placement is where the ribbon was last left; nil until it has been placed (FR-403).
	Placement *placement.Stored
	// Opacity is how opaque the ribbon's background is drawn, in percent, from MinOpacity to
	// MaxOpacity (FR-622).
	Opacity int
	// Scale is how large the cells are drawn on top of their size, in percent, from MinScale to
	// MaxScale (FR-623).
	Scale int
	// LastEdge is the edge the ribbon last stood flush against, which unpinning away from every edge
	// returns it to (FR-411, FR-613); nil until it has stood against one.
	LastEdge *placement.Against
	// PullOutSide is the ribbon's side its pull out was last on, which it keeps while it stands
	// against no edge (FR-902, Amendment 35); empty until it has had one.
	PullOutSide placement.Edge
}

// Defaults answers the choices of a first run: Classic, vertical (FR-103), the system's theme, not on
// top, pinned (FR-613), not yet placed, wholly opaque at its own size.
func Defaults() Choices {
	return Choices{
		Colour:      Classic,
		Orientation: Vertical,
		Theme:       System,
		Pinned:      true,
		Opacity:     MaxOpacity,
		Scale:       WholeScale,
	}
}

// PinnedInEffect answers whether the ribbon behaves as pinned, flush telling whether it stands flush
// against an edge running along its orientation: pinned when chosen so; also anywhere away from such
// an edge whatever was chosen (FR-619). The choice itself is Pinned, which this never changes.
func (c Choices) PinnedInEffect(flush bool) bool { return c.Pinned || !flush }

// OnTop answers whether the ribbon is kept above other windows, flush as for PinnedInEffect: where
// Always on top is on; always while unpinned in effect, so a tab can never be covered for good
// (FR-505, FR-617).
func (c Choices) OnTop(flush bool) bool { return c.AlwaysOnTop || !c.PinnedInEffect(flush) }

// Normalised answers the choices with any that is not one of the known values replaced by its
// default, so a hand-edited file holding a word it should not cannot leave a choice unset.
func (c Choices) Normalised() Choices {
	defaults := Defaults()
	if !slices.Contains(Colours, c.Colour) {
		c.Colour = defaults.Colour
	}
	if c.Orientation != Horizontal && c.Orientation != Vertical {
		c.Orientation = defaults.Orientation
	}
	if c.Theme != System && c.Theme != Light && c.Theme != Dark {
		c.Theme = defaults.Theme
	}
	c.Opacity = min(max(c.Opacity, MinOpacity), MaxOpacity)
	c.Scale = min(max(c.Scale, MinScale), MaxScale)
	if c.LastEdge != nil && !slices.Contains(edges, c.LastEdge.Edge) {
		c.LastEdge = nil
	}
	if !slices.Contains(edges, c.PullOutSide) {
		c.PullOutSide = ""
	}
	return c
}

// edges is every edge a ribbon can stand against.
var edges = []placement.Edge{placement.Left, placement.Right, placement.Top, placement.Bottom}
