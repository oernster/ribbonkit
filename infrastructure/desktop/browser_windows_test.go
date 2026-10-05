package desktop

import (
	"path/filepath"
	"strings"
	"testing"
)

// An address Windows has nothing to open with is refused with the address named, rather than
// ignored. Success would open a real browser, so it is checked by hand.
func TestAnAddressWindowsCannotOpenIsRefused(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "absent", "nothing-here.xyz")
	err := OpenInBrowser(absent)
	if err == nil {
		t.Fatal("an address with nothing to open it was reported opened")
	}
	if !strings.Contains(err.Error(), absent) {
		t.Errorf("the refusal %q does not name the address", err)
	}
	if err := OpenInBrowser("bad\x00address"); err == nil {
		t.Error("an address Windows cannot be handed was reported opened")
	}
}
