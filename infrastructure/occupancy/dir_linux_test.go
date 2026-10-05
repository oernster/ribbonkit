package occupancy

import (
	"errors"
	"testing"
)

// FR-412: the shared folder is ribbonkit under the user's runtime folder; none where the environment
// names none.
func TestTheSharedFolderIsUnderTheRuntimeFolder(t *testing.T) {
	t.Parallel()
	got, err := Dir(func(string) (string, bool) { return "/run/user/1000", true })
	if err != nil || got != "/run/user/1000/ribbonkit" {
		t.Errorf("got %q (%v)", got, err)
	}
	if _, err := Dir(func(string) (string, bool) { return "", false }); !errors.Is(err, ErrNoFolder) {
		t.Errorf("with none named: %v", err)
	}
}
