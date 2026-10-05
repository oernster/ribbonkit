//go:build linux || darwin

package desktop

import (
	"path/filepath"
	"strings"
	"testing"
)

// An address the desktop cannot open is refused with the address named, rather than ignored.
// Success would open a real browser, so it is checked by hand.
func TestAnAddressTheDesktopCannotOpenIsRefused(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no-such-opener")
	err := openWith(missing, "https://example.invalid/")
	if err == nil {
		t.Fatal("an address with nothing to open it was reported opened")
	}
	if !strings.Contains(err.Error(), "https://example.invalid/") {
		t.Errorf("the refusal %q does not name the address", err)
	}
	if err := OpenInBrowser("bad\x00address"); err == nil {
		t.Error("an address the desktop cannot be handed was reported opened")
	}
}
