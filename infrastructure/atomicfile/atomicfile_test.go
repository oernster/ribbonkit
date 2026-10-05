package atomicfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A write creates the file whole, then replaces it whole, leaving no temporary file behind.
func TestAWriteReplacesTheFileWhole(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	for _, body := range []string{`{"first":true}`, `{"second":true}`} {
		if err := Write(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if got, err := os.ReadFile(path); err != nil || string(got) != body {
			t.Fatalf("read %q, %v; want %q", got, err, body)
		}
	}
	if left := entries(t, dir); len(left) != 1 {
		t.Errorf("the folder holds %v; want the file alone", left)
	}
}

// A write into a folder that is not there fails, leaving nothing.
func TestAWriteWithNoFolderFails(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "missing", "settings.json")
	if err := Write(path, []byte("{}"), 0o600); err == nil || !strings.Contains(err.Error(), "writing beside") {
		t.Errorf("answered %v; want the write refused", err)
	}
}

// A file that cannot be replaced (here a folder stands at its name) is left as it was and the
// temporary file is removed.
func TestAFailedReplaceLeavesNoTemporaryFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "inside"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, []byte("{}"), 0o600); err == nil || !strings.Contains(err.Error(), "replacing") {
		t.Fatalf("answered %v; want the replace refused", err)
	}
	if left := entries(t, dir); len(left) != 1 || left[0] != "settings.json" {
		t.Errorf("the folder holds %v; want the folder alone, no temporary file", left)
	}
}

// entries answers the names in dir.
func entries(t *testing.T, dir string) []string {
	t.Helper()
	found, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, each := range found {
		names = append(names, each.Name())
	}
	return names
}
