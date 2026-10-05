package settingsfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// byteOrderMark is UTF-8's, which Notepad and PowerShell 5 can put before the text.
const byteOrderMark = "\xef\xbb\xbf"

// indent is the indentation the file is written with.
const indent = "  "

// Object is a settings file's top-level object: each key with its value as written.
type Object map[string]json.RawMessage

// member is one top-level key with its value as written.
type member struct {
	key   string
	value json.RawMessage
}

// parse reads raw, less any byte order mark, as an object; false when it is not one.
func parse(raw []byte) (Object, bool) {
	var object Object
	if err := json.Unmarshal(trimmed(raw), &object); err != nil || object == nil {
		return nil, false
	}
	return object, true
}

// trimmed answers raw without a leading byte order mark.
func trimmed(raw []byte) []byte { return bytes.TrimPrefix(raw, []byte(byteOrderMark)) }

// Read decodes object[key] into target, leaving target as it was when the key is missing or its
// value is the wrong type, so one bad value leaves its default and the rest load.
func Read[T any](object Object, key string, target *T) {
	value, ok := object[key]
	if !ok {
		return
	}
	var read T
	if json.Unmarshal(value, &read) == nil {
		*target = read
	}
}

// List answers object[key] as a list of entries, each as written. A missing key or a null is no
// entries; false when the value is there but is not a list, since then nothing in it can be trusted.
func (o Object) List(key string) ([]json.RawMessage, bool) {
	value, ok := o[key]
	if !ok || string(value) == "null" {
		return nil, true
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(value, &entries); err != nil {
		return nil, false
	}
	return entries, true
}

// extrasOf answers the top-level keys not in known, in the order the file has them.
func extrasOf(raw []byte, object Object, known []string) []member {
	var extras []member
	decoder := json.NewDecoder(bytes.NewReader(trimmed(raw)))
	if _, err := decoder.Token(); err != nil {
		return nil
	}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return extras
		}
		key, _ := token.(string)
		var skipped json.RawMessage
		if decoder.Decode(&skipped) != nil {
			return extras
		}
		if !slices.Contains(known, key) {
			extras = append(extras, member{key: key, value: object[key]})
		}
	}
	return extras
}

// encode answers the file's text: keys in order with their values, then the keys a later version
// wrote, each as found, indented and ending in a newline.
func encode(keys []string, values map[string]any, extras []member) ([]byte, error) {
	var compact bytes.Buffer
	compact.WriteByte('{')
	for index, key := range keys {
		value, ok := values[key]
		if !ok {
			return nil, fmt.Errorf("writing %s: no value was given for it", key)
		}
		written, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("writing %s: %w", key, err)
		}
		writeMember(&compact, index > 0, key, written)
	}
	for _, extra := range extras {
		writeMember(&compact, true, extra.key, extra.value)
	}
	compact.WriteByte('}')
	var out bytes.Buffer
	if err := json.Indent(&out, compact.Bytes(), "", indent); err != nil {
		return nil, fmt.Errorf("indenting the settings: %w", err)
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

// writeMember writes key and value to buffer, after a comma where one is owed.
func writeMember(buffer *bytes.Buffer, comma bool, key string, value []byte) {
	if comma {
		buffer.WriteByte(',')
	}
	name, _ := json.Marshal(key)
	buffer.Write(name)
	buffer.WriteByte(':')
	buffer.Write(value)
}
