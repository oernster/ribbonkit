package structural

// The layer rules know a file's layer by the folder it sits in. The folders that are not layers are
// named here: web (the page's half), installer (the setup program's window over the install
// policy, a program's composition rather than a layer), structure (the structural tests' mechanics),
// tests and tools (the generators an application runs by hand, which hold no Go). These tests hold
// the kit to those folders alone, so a new one cannot escape the layer rules unnoticed. They also
// hold the setup program to the one package of the module it is the window over.

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// kitFolders are the folders the kit may hold: its four layers, the page's half, the setup
// program, the structural mechanics, the tests and the hand-run generators.
var kitFolders = []string{
	structure.Application, structure.Domain, structure.Infrastructure, structure.UI,
	"web", "installer", "structure", "tests", generators,
}

// generators is the folder of the generators an application runs by hand; it holds no Go, so
// nothing there can sit outside the layer rules.
const generators = "tools"

// unheldFolders are folders a working copy holds that are not the kit's: git's, npm's and Claude's.
var unheldFolders = []string{".git", "node_modules", ".claude"}

// setupProgram is the setup program's folder; installPolicy is the one package of the module it
// may import.
var (
	setupProgram  = "installer"
	installPolicy = module + "/infrastructure/setup"
)

func TestTheKitHoldsOnlyItsLayersThePageAndTheSetupProgram(t *testing.T) {
	entries, err := os.ReadDir(structure.Root(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && !slices.Contains(kitFolders, name) && !slices.Contains(unheldFolders, name) {
			t.Errorf("%s is neither a layer nor one of the kit's named folders: the layer rules cannot see it", name)
		}
	}
}

func TestTheGeneratorsHoldNoGo(t *testing.T) {
	root := structure.Root(t)
	for _, path := range goFiles(t) {
		if relative := structure.Relative(root, path); strings.HasPrefix(relative, generators+"/") {
			t.Errorf("%s is Go under %s, where the layer rules do not look", relative, generators)
		}
	}
}

func TestTheSetupProgramReachesOnlyTheInstallPolicy(t *testing.T) {
	root := structure.Root(t)
	read := 0
	for _, path := range goFiles(t) {
		relative := structure.Relative(root, path)
		if !strings.HasPrefix(relative, setupProgram+"/") {
			continue
		}
		read++
		for _, imported := range structure.ImportsOf(t, path) {
			if strings.HasPrefix(imported, module+"/") && imported != installPolicy {
				t.Errorf("%s imports %s: the setup program is the window over the install policy alone", relative, imported)
			}
		}
	}
	if read == 0 {
		t.Fatalf("no Go file found under %s, the walk is wrong", filepath.ToSlash(setupProgram))
	}
}
