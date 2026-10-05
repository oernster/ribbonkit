package occupancy

import (
	"errors"
	"testing"
)

// FR-412: the shared folder is ribbonkit under the local application data folder; none where the
// environment names none.
func TestTheSharedFolderIsUnderLocalAppData(t *testing.T) {
	t.Parallel()
	got, err := Dir(func(string) (string, bool) { return `C:\Users\Someone\AppData\Local`, true })
	if err != nil || got != `C:\Users\Someone\AppData\Local\ribbonkit` {
		t.Errorf("got %q (%v)", got, err)
	}
	if _, err := Dir(func(string) (string, bool) { return "", false }); !errors.Is(err, ErrNoFolder) {
		t.Errorf("with none named: %v", err)
	}
}
