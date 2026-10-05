package heldfile

import (
	"os"
	"path/filepath"
	"testing"
)

// While a file is held nobody else can open or remove it; once let go, both work again.
func TestAHeldFileCanBeNeitherOpenedNorRemoved(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	release, err := Hold(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(path); err == nil {
		t.Error("a held file was read")
	}
	if err := os.Remove(path); err == nil {
		t.Error("a held file was removed")
	}
	release()
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("a released file could not be read: %v", err)
	}
}

// Neither a file that is not there nor a path no file can have can be held.
func TestWhatCannotBeOpenedCannotBeHeld(t *testing.T) {
	t.Parallel()
	for _, path := range []string{filepath.Join(t.TempDir(), "missing"), "settings\x00.json"} {
		if release, err := Hold(path); err == nil {
			release()
			t.Errorf("%q was held", path)
		}
	}
}
