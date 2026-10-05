//go:build windows

package installer

import (
	"testing"
	"testing/fstest"
)

// Pictures under a root that cannot be opened are refused before any window is shown, naming what
// was being read.
func TestPicturesThatCannotBeOpenedAreRefused(t *testing.T) {
	t.Parallel()
	err := show(nil, Program{Pictures: fstest.MapFS{}, PicturesRoot: "../outside"})
	if err == nil {
		t.Fatal("pictures under a root outside their folder were accepted")
	}
}
