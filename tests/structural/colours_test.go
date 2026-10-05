package structural

// The menus offer the schemes the ribbon domain lists; the kit's half of the palette (the ribbon's
// own tokens) must state each one in full. An application holds its own half to the same rule and
// checks the contrast of the two halves together, since only it knows what its words are drawn on.

import (
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/ribbonkit/structure"
)

// optionalTokens may be left to Classic's value: its problem colour already meets the contrast floor
// on every scheme's cell and surface (TimeRibbon NFR-U-1).
var optionalTokens = []string{"problem"}

// kitHalf is the kit's half of the palette: theme.css states Classic, colours.css every other scheme.
func kitHalf(t *testing.T) structure.Half {
	t.Helper()
	web := filepath.Join(structure.Root(t), webDir)
	return structure.Half{Classic: filepath.Join(web, "theme.css"), Schemes: filepath.Join(web, "colours.css")}
}

// offered answers the schemes the menus offer, by name.
func offered() []string {
	names := make([]string, 0, len(ribbon.Colours))
	for _, colour := range ribbon.Colours {
		names = append(names, string(colour))
	}
	return names
}

func TestEveryOfferedSchemeHasItsOwnCompleteBlock(t *testing.T) {
	structure.CheckEveryOfferedSchemeHasItsOwnCompleteBlock(t, kitHalf(t), string(ribbon.Classic), offered(), optionalTokens)
}

// The page shows the system's dark under System and the chosen dark under Dark (FR-606).
func TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t *testing.T) {
	structure.CheckClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t, kitHalf(t))
}
