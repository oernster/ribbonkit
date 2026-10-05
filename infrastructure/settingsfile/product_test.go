package settingsfile

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// The test product's own keys, placed among the ribbon's.
const (
	keyVersion = "version"
	keyName    = "name"
	keyItems   = "items"
)

// productName is the name the user knows the test product by.
const productName = "TestRibbon"

// item is one entry of the test product's list.
type item struct {
	ID, Value, Unreadable, Original string
}

// testSettings is the test product's settings: the ribbon's choices, one value and a list.
type testSettings struct {
	ribbon.Choices
	Name  string
	Items []item
}

// storedItem is one item as the file holds it. Pointers tell a missing field from an empty one.
type storedItem struct {
	ID       *string `json:"id"`
	Value    *string `json:"value"`
	Position *int    `json:"position"`
}

// errNoID is why an item without an id cannot be read.
var errNoID = errors.New("it has no id")

// items reads and marks the test product's items.
var items = Entries[item]{
	Decode: func(raw json.RawMessage) (item, *int, error) {
		var stored storedItem
		if err := json.Unmarshal(raw, &stored); err != nil {
			return item{}, nil, err
		}
		if stored.ID == nil || stored.Value == nil {
			return item{}, nil, errNoID
		}
		return item{ID: *stored.ID, Value: *stored.Value}, stored.Position, nil
	},
	Encode: func(each item, position int) (json.RawMessage, error) {
		if each.Value == unwritable {
			return nil, errUnwritable
		}
		return json.Marshal(storedItem{ID: &each.ID, Value: &each.Value, Position: &position})
	},
	Unreadable: func(id, reason, original string) item {
		return item{ID: id, Unreadable: reason, Original: original}
	},
	ID:       func(each item) (string, bool) { return each.ID, each.Unreadable == "" },
	WithID:   func(each item, id string) item { each.ID = id; return each },
	Original: func(each item) string { return each.Original },
}

// unwritable is the value an item cannot be written with, so a test can refuse a save.
const unwritable = "unwritable"

// errUnwritable is why an item holding unwritable is refused.
var errUnwritable = errors.New("the item cannot be written")

// codec is the test product's file: version, name, the ribbon's choices, then the items.
var codec = Codec[testSettings]{
	Keys: []string{
		keyVersion, keyName, KeyColour, KeyOrientation, KeyTheme, KeyAlwaysOnTop, KeyPlacement, keyItems,
		KeySkippedUpdate, KeyPinned, KeyLastEdge, KeyOpacity, KeyScale, KeyPullOutSide,
	},
	Defaults: func() testSettings { return testSettings{Choices: ribbon.Defaults(), Name: "first"} },
	Decode: func(object Object) (testSettings, bool) {
		raw, ok := object.List(keyItems)
		if !ok {
			return testSettings{}, false
		}
		decoded := testSettings{Choices: ribbon.Defaults(), Name: "first"}
		ReadChoices(object, &decoded.Choices)
		Read(object, keyName, &decoded.Name)
		decoded.Items = items.Read(raw)
		return decoded, true
	},
	Values: func(current testSettings) (map[string]any, error) {
		written, err := items.Write(current.Items)
		if err != nil {
			return nil, err
		}
		values := map[string]any{keyVersion: 1, keyName: current.Name, keyItems: written}
		maps.Copy(values, ChoiceValues(current.Choices))
		return values, nil
	},
}

// newStore answers a store of the test product over dir.
func newStore(dir string) *Store[testSettings] { return New(dir, productName, codec) }

// write puts text in dir's settings file.
func write(t *testing.T, dir, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(text), FileMode); err != nil {
		t.Fatal(err)
	}
}

// read answers dir's settings file.
func read(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// compact answers text with its insignificant whitespace removed.
func compact(t *testing.T, text string) string {
	t.Helper()
	var out strings.Builder
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, text)
	}
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
	return strings.TrimSpace(out.String())
}
