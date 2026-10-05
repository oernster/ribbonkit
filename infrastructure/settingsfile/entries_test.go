package settingsfile

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// ids answers the ids of entries in order.
func ids(entries []item) []string {
	out := make([]string, 0, len(entries))
	for _, each := range entries {
		out = append(out, each.ID)
	}
	return out
}

// One bad entry stands in as unreadable with its reason while the others load. A save writes the
// bad one back as it was found.
func TestOneBadEntryLeavesTheOthersWorking(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := `{"id": "b", "value": 7}`
	write(t, dir, `{"items": [
		{"id": "a", "value": "London", "position": 0},
		`+bad+`,
		{"id": "c", "value": "Gran", "position": 2},
		{"value": "no id"}
	]}`)
	store := newStore(dir)
	loaded, _, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := loaded.Items
	if len(got) != 4 || got[0].Value != "London" || got[2] != (item{ID: "c", Value: "Gran"}) {
		t.Fatalf("items %+v", got)
	}
	for index, reason := range map[int]string{1: "cannot unmarshal", 3: errNoID.Error()} {
		if !strings.Contains(got[index].Unreadable, reason) || !strings.HasPrefix(got[index].ID, UnreadableIDPrefix) {
			t.Errorf("entry %d: %+v", index, got[index])
		}
	}
	if err := store.Save(loaded); err != nil {
		t.Fatal(err)
	}
	if text := compact(t, read(t, dir)); !strings.Contains(text, compact(t, bad)) {
		t.Errorf("the bad entry was not written back as found:\n%s", text)
	}
}

// The stored position keeps the order entries were added in; an entry with none keeps its place.
func TestOrderingPersistsByPosition(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"items": [
		{"id": "late", "value": "x", "position": 2},
		{"id": "none", "value": "x"},
		{"id": "early", "value": "x", "position": 0}
	]}`)
	loaded, _, _ := newStore(dir).Load()
	if got := ids(loaded.Items); !slices.Equal(got, []string{"early", "none", "late"}) {
		t.Errorf("got %v", got)
	}
}

// Two entries sharing an id are both kept, each answering to an id of its own, which a save writes.
func TestEntriesSharingAnIDAreToldApart(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"items": [
		{"id": "same", "value": "London"},
		{"id": "same", "value": "Tokyo"},
		{"id": "same-2", "value": "Paris"}
	]}`)
	store := newStore(dir)
	loaded, _, err := store.Load()
	if got := ids(loaded.Items); err != nil || !slices.Equal(got, []string{"same", "same-3", "same-2"}) {
		t.Fatalf("ids %v (%v)", got, err)
	}
	if err := store.Save(loaded); err != nil {
		t.Fatal(err)
	}
	if again, _, _ := newStore(dir).Load(); !slices.Equal(ids(again.Items), ids(loaded.Items)) {
		t.Errorf("the ids given were not saved: %v", ids(again.Items))
	}
}

// An unreadable entry's id for the session never collides with an id the file holds.
func TestAnUnreadableEntryNeverTakesAWrittenID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	written := UnreadableIDPrefix + "1"
	write(t, dir, `{"items": [{"id": "`+written+`", "value": "x"}, {"id": 7}]}`)
	loaded, _, err := newStore(dir).Load()
	got := ids(loaded.Items)
	if err != nil || len(got) != 2 || got[0] != written || got[1] == written {
		t.Errorf("ids %v (%v)", got, err)
	}
}

// An unreadable entry found as JSON is written back as found; one whose text is not JSON is written
// as any other entry. An entry that cannot be written refuses the whole list.
func TestWhatIsWrittenBackAndWhatIsRefused(t *testing.T) {
	t.Parallel()
	written, err := items.Write([]item{
		{ID: "unreadable-0", Unreadable: "bad", Original: `{"id": 7}`},
		{ID: "unreadable-1", Unreadable: "bad", Original: `{"id": `},
		{ID: "a", Value: "x"},
	})
	if err != nil || len(written) != 3 || string(written[0]) != `{"id": 7}` ||
		string(written[1]) != `{"id":"unreadable-1","value":"","position":1}` ||
		string(written[2]) != `{"id":"a","value":"x","position":2}` {
		t.Errorf("wrote %s (%v)", written, err)
	}
	if _, err := items.Write([]item{{ID: "a", Value: unwritable}}); !errors.Is(err, errUnwritable) {
		t.Errorf("answered %v", err)
	}
}
