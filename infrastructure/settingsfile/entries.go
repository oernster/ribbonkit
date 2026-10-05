package settingsfile

import (
	"encoding/json"
	"slices"
	"strconv"
)

// UnreadableIDPrefix begins the id an unreadable entry is given for the session, so it can be
// removed. It is never written: the entry is written back as it was found.
const UnreadableIDPrefix = "unreadable-"

// firstNumberedID is the number the first id made from another is given; the original is the first.
const firstNumberedID = 2

// Entries is how one application reads and writes the entries of a list: its clocks, its cities.
type Entries[E any] struct {
	// Decode reads one entry, answering it with its stored position (nil where it has none) or the
	// reason it cannot be read, which the user is shown.
	Decode func(raw json.RawMessage) (E, *int, error)
	// Encode answers a read entry as written at position.
	Encode func(entry E, position int) (json.RawMessage, error)
	// Unreadable answers the entry standing in for one that could not be read: its id for the
	// session, the reason and the text found, which is written back.
	Unreadable func(id, reason, original string) E
	// ID answers an entry's id and whether the entry was read.
	ID func(E) (string, bool)
	// WithID answers the entry holding id.
	WithID func(E, string) E
	// Original answers the text an unreadable entry was found as.
	Original func(E) string
}

// Write answers each entry as written, in order. An entry that could not be read is written back
// exactly as it was found, so nothing the user wrote is lost; where that text is not JSON it is
// written as any other entry.
func (e Entries[E]) Write(entries []E) ([]json.RawMessage, error) {
	written := make([]json.RawMessage, 0, len(entries))
	for position, entry := range entries {
		_, read := e.ID(entry)
		if raw, ok := writtenBack(e.Original(entry)); !read && ok {
			written = append(written, raw)
			continue
		}
		raw, err := e.Encode(entry, position)
		if err != nil {
			return nil, err
		}
		written = append(written, raw)
	}
	return written, nil
}

// Read reads each entry on its own, so one bad entry stands in as unreadable and the rest load.
// Entries are ordered by their stored position; an entry with none keeps its place in the list.
// Every id is then made the entry's own, since a change names its entry by id.
func (e Entries[E]) Read(raw []json.RawMessage) []E {
	type ordered struct {
		entry    E
		position int
	}
	read := make([]ordered, 0, len(raw))
	for index, each := range raw {
		entry, position, err := e.Decode(each)
		if err != nil {
			entry, position = e.Unreadable(UnreadableIDPrefix+strconv.Itoa(index), err.Error(), string(each)), nil
		}
		at := index
		if position != nil {
			at = *position
		}
		read = append(read, ordered{entry: entry, position: at})
	}
	slices.SortStableFunc(read, func(a, b ordered) int { return a.position - b.position })
	out := make([]E, 0, len(read))
	for _, each := range read {
		out = append(out, each.entry)
	}
	return e.withUniqueIDs(out)
}

// withUniqueIDs answers entries with every id its own. The first read entry holding an id keeps it;
// a later one holding it again, as a hand-edited file can, is given that id numbered, which a save
// then writes. An unreadable entry's id is the session's alone, so it gives way to any id the file
// holds.
func (e Entries[E]) withUniqueIDs(entries []E) []E {
	taken := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if id, read := e.ID(entry); read {
			taken[id] = true
		}
	}
	kept := make(map[string]bool, len(entries))
	for index, entry := range entries {
		id, read := e.ID(entry)
		if read && !kept[id] {
			kept[id] = true
			continue
		}
		free := freeID(id, taken)
		entries[index] = e.WithID(entry, free)
		taken[free] = true
	}
	return entries
}

// freeID answers id numbered with the first number that makes it one taken does not hold.
func freeID(id string, taken map[string]bool) string {
	for n := firstNumberedID; ; n++ {
		candidate := id + "-" + strconv.Itoa(n)
		if !taken[candidate] {
			return candidate
		}
	}
}

// writtenBack answers the text an unreadable entry was found as, to be written back unchanged; false
// where that text is not JSON, so the entry must be written some other way.
func writtenBack(original string) (json.RawMessage, bool) {
	if !json.Valid([]byte(original)) {
		return nil, false
	}
	return json.RawMessage(original), true
}
