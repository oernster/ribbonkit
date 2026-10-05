package structural

// The menus offer the schemes the ribbon domain lists; the kit's half of the palette (the ribbon's
// own tokens) must state each one in full, its words legible on what it draws them on. An
// application holds its own half to the same rules and checks the contrast of the two halves
// together, since only it knows what its words are drawn on.

import (
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/ribbonkit/structure"
)

// optionalTokens may be left to Classic's value: its problem colour already meets the contrast floor
// on every scheme's cell and surface (TimeRibbon NFR-U-1).
var optionalTokens = []string{"problem"}

// kitHalf is the kit's own half of the palette.
func kitHalf(t *testing.T) structure.Half {
	t.Helper()
	return structure.KitHalf(structure.Root(t))
}

func TestEveryOfferedSchemeHasItsOwnCompleteBlock(t *testing.T) {
	structure.CheckEveryOfferedSchemeHasItsOwnCompleteBlock(t, kitHalf(t), string(ribbon.Classic), structure.Offered(), optionalTokens)
}

func TestTheKitsTextMeetsTheContrastFloor(t *testing.T) {
	structure.CheckTextMeetsTheContrastFloor(t, []structure.Half{kitHalf(t)}, structure.Offered(), structure.TextTokens(), structure.Backgrounds())
}

// The page shows the system's dark under System and the chosen dark under Dark (FR-606).
func TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t *testing.T) {
	structure.CheckClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t, kitHalf(t))
}
