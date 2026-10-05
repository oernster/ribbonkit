package structural

// A ribbon's one network request is its update check (TimeRibbon NFR-S-1, FR-509). In the kit,
// only the update check's adapter may import a network package, only the files named here may
// start a program and neither page half may ask a network.

import (
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// networkExempt is the one directory whose files may import a network package: the update check's
// adapter, which asks GitHub for the latest release.
const networkExempt = "infrastructure/update"

// processStarters are the files that may start a program or hand an address to the desktop. Every
// one hands work to the desktop or to setup; none asks a network.
var processStarters = []string{
	"infrastructure/desktop/browser_unix.go",    // hands an address to open or xdg-open
	"infrastructure/desktop/browser_windows.go", // hands an address to ShellExecute
	"infrastructure/setup/deletion.go",          // deletes the install folder after setup exits
	"infrastructure/setup/process.go",           // starts the installed application
}

func TestOnlyTheUpdateCheckImportsANetworkPackage(t *testing.T) {
	structure.CheckOnlyTheExemptImportANetworkPackage(t, structure.Root(t), goFiles(t), networkExempt)
}

func TestOnlyNamedFilesStartAProcess(t *testing.T) {
	structure.CheckOnlyNamedFilesStartAProcess(t, structure.Root(t), goFiles(t), processStarters)
}

func TestNeitherPageHalfMakesARequest(t *testing.T) {
	structure.CheckThePageMakesNoRequest(t, append(webFiles(t), setupPageFiles(t)...))
}
