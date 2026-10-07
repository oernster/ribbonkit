package structural

import (
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

func TestDomainHasNoOutwardImports(t *testing.T) {
	kitLayout(t).CheckDomainHasNoOutwardImports(t, goFiles(t))
}

func TestDomainIsPure(t *testing.T) {
	kitLayout(t).CheckDomainIsPure(t, goFiles(t))
}

func TestApplicationDoesNotImportInfrastructure(t *testing.T) {
	kitLayout(t).CheckApplicationDoesNotImportInfrastructure(t, goFiles(t))
}

// The ui layer reaches the desktop through shell.Desktop, never the desktop package itself.
func TestTheUIDependsOnTheApplicationOnly(t *testing.T) {
	kitLayout(t).CheckUIDependsOnTheApplicationOnly(t, goFiles(t))
}

func TestNothingBelowTheUIImportsIt(t *testing.T) {
	kitLayout(t).CheckNothingBelowTheUIImportsIt(t, goFiles(t))
}

func TestWailsStaysOutOfInfrastructure(t *testing.T) {
	kitLayout(t).CheckWailsStaysOutOfInfrastructure(t, goFiles(t))
}

// The kit has no composition root: each application wires the kit's adapters in its own.
func TestNothingWiresTheApplicationToTheInfrastructure(t *testing.T) {
	kitLayout(t).CheckCompositionRootIsWhitelisted(t, goFiles(t))
}

// sourceFiles is every file the size rule governs: the Go, its C and Objective-C halves, the web
// half and the setup page.
func sourceFiles(t *testing.T) []string {
	t.Helper()
	files := append(goFiles(t), nativeFiles(t)...)
	return append(append(files, webFiles(t)...), setupPageFiles(t)...)
}

func TestNoFileExceedsLineLimit(t *testing.T) {
	structure.CheckNoFileExceedsLineLimit(t, sourceFiles(t))
}

func TestNoFileInDangerBand(t *testing.T) {
	structure.CheckNoFileInDangerBand(t, sourceFiles(t))
}

func TestEveryExportedTypeIsDocumented(t *testing.T) {
	structure.CheckEveryExportedTypeIsDocumented(t, goFiles(t))
}
