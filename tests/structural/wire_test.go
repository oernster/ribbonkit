package structural

// The kit's half of the wire: the window's Go types in ui/window/wire.go, stated again as
// TypeScript in web/wire.ts; also the words the window and the setup facade send their pages.

import (
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// wirePairs names each Go wire type with the TypeScript interface stating it again.
var wirePairs = map[string]string{
	"aboutDTO": "AboutFacts", "creditDTO": "Credit", "updateDTO": "UpdateStatus", "Box": "Box",
}

// The files each side of the wire and its words are stated in, from the repository's root.
var (
	goWireFile  = filepath.Join("ui", "window", "wire.go")
	tsWireFile  = filepath.Join(webDir, "wire.ts")
	windowFile  = filepath.Join("ui", "window", "facade.go")
	setupFacade = filepath.Join("installer", "facade.go")
)

// windowWords names each constant the window sends the page with the shape the page must state its
// value in: a listener for an event and a key of panelFor for a panel. setupWords does the same for
// the setup facade and the setup page.
var (
	windowWords = map[string]string{
		"eventRefresh":   heardByTheWindow,
		"eventOpenPanel": heardByTheWindow,
		"openAtSettings": panelKey,
		"openAtAbout":    panelKey,
		"openAtLicence":  panelKey,
		"openAtUpdate":   panelKey,
	}
	setupWords = map[string]string{"progressEvent": "EventsOn('%s'"}
)

const (
	heardByTheWindow = "on('%s'"
	panelKey         = "'%s': "
)

func TestTheWireIsStatedAlikeOnBothSides(t *testing.T) {
	root := structure.Root(t)
	structure.CheckTheWireIsStatedAlikeOnBothSides(t, wirePairs,
		[]string{filepath.Join(root, goWireFile)}, []string{filepath.Join(root, tsWireFile)})
}

func TestThePageNamesEveryEventTheWindowEmits(t *testing.T) {
	structure.CheckThePageNamesEveryWord(t, filepath.Join(structure.Root(t), windowFile), windowWords, webFiles(t))
}

// The setup program's progress bar moves only on the word the setup facade emits.
func TestTheSetupPageNamesEveryEventSetupEmits(t *testing.T) {
	structure.CheckThePageNamesEveryWord(t, filepath.Join(structure.Root(t), setupFacade), setupWords, setupPageFiles(t))
}
