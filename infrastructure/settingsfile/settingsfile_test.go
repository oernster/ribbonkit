package settingsfile

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// threeItems is a file holding three items, the user's own.
const threeItems = `{"items": [
	{"id": "m", "value": "Mum"},
	{"id": "k", "value": "Kenji"},
	{"id": "o", "value": "Office"}
]}`

// No file is the defaults with no notice; loading makes no folder.
func TestAbsentFileMeansDefaults(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), productName)
	loaded, notice, err := newStore(dir).Load()
	if err != nil || notice != "" || loaded.Name != "first" || len(loaded.Items) != 0 {
		t.Errorf("got %+v %q (%v)", loaded, notice, err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Error("loading made the folder")
	}
}

// A file that cannot be trusted is kept aside as it was, the defaults stand in and the user is told
// where it went.
func TestUnreadableFileIsKeptAsideAndReported(t *testing.T) {
	t.Parallel()
	for name, text := range map[string]string{"not JSON": "{ this is not", "items not a list": `{"items": 5}`, "a list": `[1]`} {
		dir := t.TempDir()
		write(t, dir, text)
		loaded, notice, err := newStore(dir).Load()
		if err != nil || notice != KeptAsideNotice(UnreadableName) || loaded.Name != "first" {
			t.Errorf("%s: got %+v %q (%v)", name, loaded, notice, err)
		}
		kept, err := os.ReadFile(filepath.Join(dir, UnreadableName))
		if err != nil || string(kept) != text {
			t.Errorf("%s: kept %q (%v)", name, kept, err)
		}
	}
}

// A second damaged file never replaces the copy kept aside from the first.
func TestASecondDamagedFileKeepsTheFirstCopy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	first := strings.TrimSuffix(threeItems, "]}")
	write(t, dir, first)
	if _, notice, err := newStore(dir).Load(); err != nil || notice != KeptAsideNotice(UnreadableName) {
		t.Fatalf("first: %q (%v)", notice, err)
	}
	second := `{"items": [ PARTIAL`
	write(t, dir, second)
	_, notice, err := newStore(dir).Load()
	secondName := KeptAsideName(2)
	if err != nil || notice != KeptAsideNotice(secondName) {
		t.Errorf("second: %q (%v)", notice, err)
	}
	for name, want := range map[string]string{UnreadableName: first, secondName: second} {
		kept, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(kept) != want {
			t.Errorf("%s holds %q (%v), want %q", name, kept, err, want)
		}
	}
}

// When the unreadable file cannot be kept aside, every name being taken, nothing overwrites it.
func TestAFileThatCannotBeKeptAsideIsNeverOverwritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, "{ not JSON")
	for n := 1; n <= KeptAsideLimit; n++ {
		if err := os.Mkdir(filepath.Join(dir, KeptAsideName(n)), FolderMode); err != nil {
			t.Fatal(err)
		}
	}
	store := newStore(dir)
	if _, _, err := store.Load(); !errors.Is(err, ErrNotKeptAside) {
		t.Errorf("load answered %v", err)
	}
	if err := store.Save(codec.Defaults()); !errors.Is(err, ErrNotKeptAside) {
		t.Errorf("save answered %v", err)
	}
	if read(t, dir) != "{ not JSON" {
		t.Error("the unreadable file was overwritten")
	}
}

// A file that is there but cannot be read is never saved over, even once it can be read again: the
// defaults standing in for it are not the user's. The refusal names the product to start again.
func TestAFileThatCouldNotBeReadIsNeverSavedOver(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, FileName)
	if err := os.Mkdir(blocker, FolderMode); err != nil {
		t.Fatal(err)
	}
	store := newStore(dir)
	_, _, err := store.Load()
	if !errors.Is(err, ErrNotRead) || !strings.HasSuffix(err.Error(), "; nothing is saved over it until "+productName+" is started again and reads it") {
		t.Errorf("load answered %v", err)
	}
	if err := os.Remove(blocker); err != nil {
		t.Fatal(err)
	}
	write(t, dir, threeItems)
	if err := store.Save(codec.Defaults()); !errors.Is(err, ErrNotRead) {
		t.Errorf("save answered %v", err)
	}
	if read(t, dir) != threeItems {
		t.Error("the file that could not be read was saved over")
	}
}

// A replacement that fails leaves no temporary file behind.
func TestAFailedReplacementLeavesNoTemporaryFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, FileName, "inside"), FolderMode); err != nil {
		t.Fatal(err)
	}
	if err := newStore(dir).Save(codec.Defaults()); err == nil {
		t.Error("replacing a folder succeeded")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("the folder holds %v", entries)
	}
}

// A save replaces the file whole, indented in writing order, leaving nothing else behind.
func TestWriteReplacesAtomically(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"name": "old"}`)
	if err := newStore(dir).Save(codec.Defaults()); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != FileName {
		t.Errorf("the folder holds %v", entries)
	}
	if !strings.HasPrefix(read(t, dir), "{\n  \"version\": 1,\n  \"name\": \"first\",\n  \"colour\":") {
		t.Errorf("not indented in writing order:\n%s", read(t, dir))
	}
}

// A folder that cannot be made is answered, not ignored.
func TestAFolderThatCannotBeMadeIsReported(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, FileMode); err != nil {
		t.Fatal(err)
	}
	if err := newStore(filepath.Join(blocker, productName)).Save(codec.Defaults()); err == nil || !strings.Contains(err.Error(), "making") {
		t.Errorf("saving under a file answered %v", err)
	}
}

// A value with no key, a key with no value and a value JSON cannot hold are each refused, writing
// nothing.
func TestAValueThatCannotBeWrittenIsRefused(t *testing.T) {
	t.Parallel()
	for name, values := range map[string]func(testSettings) (map[string]any, error){
		"a fault":       func(testSettings) (map[string]any, error) { return nil, errNoID },
		"no value":      func(testSettings) (map[string]any, error) { return map[string]any{}, nil },
		"not JSON-able": func(testSettings) (map[string]any, error) { return map[string]any{keyVersion: math.NaN()}, nil },
	} {
		dir := t.TempDir()
		broken := codec
		broken.Values = values
		if err := New(dir, productName, broken).Save(codec.Defaults()); err == nil {
			t.Errorf("%s: saved", name)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Errorf("%s: the folder holds %v", name, entries)
		}
	}
}

// A later version's keys survive this version saving, after the known ones in their order.
func TestUnknownKeysAreKeptOnWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"version": 2, "future": {"x": [1, 2]}, "name": "mine", "later": true}`)
	store := newStore(dir)
	loaded, _, _ := store.Load()
	if !slices.Equal(store.Extras(), []string{"future", "later"}) {
		t.Errorf("extras %v", store.Extras())
	}
	if err := store.Save(loaded); err != nil {
		t.Fatal(err)
	}
	text := read(t, dir)
	var object map[string]any
	if err := json.Unmarshal([]byte(text), &object); err != nil {
		t.Fatal(err)
	}
	if object["later"] != true || !strings.Contains(compact(t, text), `"future":{"x":[1,2]}`) || object[keyName] != "mine" {
		t.Errorf("a key was lost:\n%s", text)
	}
	last, future, later := strings.Index(text, `"`+KeyPullOutSide+`"`), strings.Index(text, `"future"`), strings.Index(text, `"later"`)
	if !(last < future && future < later) {
		t.Errorf("unknown keys are not written after the known ones in their order:\n%s", text)
	}
}

// A file saved with a UTF-8 byte order mark, as Notepad and PowerShell 5 can, reads whole.
func TestAByteOrderMarkIsRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, byteOrderMark+`{"name": "marked", "later": 1, "items": [{"id": "a", "value": "x"}]}`)
	store := newStore(dir)
	loaded, notice, err := store.Load()
	if err != nil || notice != "" || loaded.Name != "marked" || len(loaded.Items) != 1 {
		t.Fatalf("loaded %+v %q (%v)", loaded, notice, err)
	}
	if err := store.Save(loaded); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, dir), `"later": 1`) {
		t.Errorf("an unknown key was lost:\n%s", read(t, dir))
	}
}
