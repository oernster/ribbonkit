package occupancy

import (
	"errors"
	"testing"
)

// FR-412: the shared folder is ribbonkit in Application Support; none where the environment names no
// home.
func TestTheSharedFolderIsInApplicationSupport(t *testing.T) {
	t.Parallel()
	got, err := Dir(func(string) (string, bool) { return "/Users/someone", true })
	if err != nil || got != "/Users/someone/Library/Application Support/ribbonkit" {
		t.Errorf("got %q (%v)", got, err)
	}
	if _, err := Dir(func(string) (string, bool) { return "", false }); !errors.Is(err, ErrNoFolder) {
		t.Errorf("with none named: %v", err)
	}
}
