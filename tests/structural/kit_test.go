// Package structural holds the kit to its architecture with tests rather than convention. The
// mechanics live in the structure package, which every repository built on the kit shares; the
// rules' data here is the kit's own.
//
// Every assertion here has been proved to bite by planting a violation and watching it fail. An
// assertion never seen to fail is not yet a guard.
package structural

import (
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// module is the kit's module path, as go.mod states it.
const module = "github.com/oernster/ribbonkit"

// The kit's two page halves: the web half an application's page is built from; the setup
// program's page, which has no build step, so its files are the source rather than an output.
var (
	webDir          = "web"
	setupPageDir    = filepath.Join("installer", "page")
	webExtensions   = []string{".ts", ".tsx", ".css"}
	setupExtensions = []string{".html", ".css", ".js"}
)

// The C and Objective-C halves of the infrastructure, written by hand beside the Go that calls
// them; they are source like any other, so the line caps reach them too.
var (
	nativeDir        = "infrastructure"
	nativeExtensions = []string{".c", ".m", ".h"}
)

// kitLayout answers the kit's layout: its layers sit at the repository's root and it is laid out
// on no other module. Nothing in it wires the application to the infrastructure.
func kitLayout(t *testing.T) structure.Layout {
	t.Helper()
	return structure.Layout{Root: structure.Root(t), Module: module, Layered: []string{"."}}
}

// goFiles answers every Go file of the kit.
func goFiles(t *testing.T) []string {
	t.Helper()
	return structure.GoFiles(t, structure.Root(t))
}

// webFiles answers the web half's source files.
func webFiles(t *testing.T) []string {
	t.Helper()
	return structure.FilesWith(t, webExtensions, filepath.Join(structure.Root(t), webDir))
}

// nativeFiles answers the infrastructure's C and Objective-C source files.
func nativeFiles(t *testing.T) []string {
	t.Helper()
	return structure.FilesWith(t, nativeExtensions, filepath.Join(structure.Root(t), nativeDir))
}

// setupPageFiles answers the setup page's own source files; the images beside them are artwork.
func setupPageFiles(t *testing.T) []string {
	t.Helper()
	return structure.FilesWith(t, setupExtensions, filepath.Join(structure.Root(t), setupPageDir))
}
