// Package appdata answers the folder an application keeps its settings and log in (FR-701,
// NFR-O-1), named by its identity: %APPDATA%\<name> on Windows, <name> in Application Support on
// macOS, <name> in the XDG configuration folder on Linux. Nothing is made here; the store and the log
// make the folder when they write. Each platform supplies where the folder's parent is in a file of
// its own.
package appdata

import (
	"errors"
	"path/filepath"

	"github.com/oernster/ribbonkit/domain/identity"
)

// ErrNoAppData is answered when the environment names no application data folder.
var ErrNoAppData = errors.New("the environment names no folder for application data")

// Dir answers app's folder, reading the environment through lookup.
func Dir(app identity.App, lookup func(string) (string, bool)) (string, error) {
	parent, ok := base(lookup)
	if !ok {
		return "", ErrNoAppData
	}
	return filepath.Join(parent, app.Name), nil
}
