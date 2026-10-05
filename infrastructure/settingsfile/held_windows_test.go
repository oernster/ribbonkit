package settingsfile

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/infrastructure/heldfile"
)

// As it happens at sign-in: the file is held open at launch, then let go before the first save. The
// settings in it are never replaced by the defaults.
func TestAFileHeldOpenAtLaunchIsNeverSavedOver(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, threeItems)
	store := newStore(dir)
	release, err := heldfile.Hold(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.Load()
	release()
	if !errors.Is(err, ErrNotRead) {
		t.Errorf("load answered %v", err)
	}
	if err := store.Save(codec.Defaults()); !errors.Is(err, ErrNotRead) {
		t.Errorf("save answered %v", err)
	}
	if read(t, dir) != threeItems {
		t.Error("the settings were replaced by the defaults")
	}
}
