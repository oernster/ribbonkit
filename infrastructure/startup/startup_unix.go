//go:build linux || darwin

package startup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/oernster/ribbonkit/domain/identity"
)

// Off Windows the entry is one file in a folder the session reads at sign-in: an XDG autostart
// entry on Linux, a launchd agent on macOS. Reading, writing and removing it is the same on both;
// each platform supplies the folder, the file's suffix, its text and what marks it switched off.

const (
	folderMode = fs.FileMode(0o700)
	entryMode  = fs.FileMode(0o644)
)

// errNoHome is answered when the session names no home folder, so there is nowhere to write.
var errNoHome = errors.New("the session names no home folder, so there is no folder for sign-in entries")

// Entry is the application's StartupEntry port over one sign-in entry file.
type Entry struct {
	app     identity.App
	dir     string
	command string
	problem error
}

// Command answers what the entry runs: the program with no arguments, so a sign-in start shows the
// ribbon as a normal launch does (FR-605).
func (e Entry) Command() string { return e.command }

// path is the entry's file, named by the app id as both platforms' conventions have it.
func (e Entry) path() string { return filepath.Join(e.dir, e.app.AppID+entrySuffix) }

// Enabled answers whether the entry is present and not switched off. A missing entry is off; one
// that is there and cannot be read is a fault, answered as one.
func (e Entry) Enabled() (bool, error) {
	if e.problem != nil {
		return false, e.problem
	}
	raw, err := os.ReadFile(e.path())
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", e.path(), err)
	}
	return !switchedOff(string(raw)), nil
}

// Enable writes the entry, making its folder where it is missing.
func (e Entry) Enable() error {
	if e.problem != nil {
		return e.problem
	}
	if err := os.MkdirAll(e.dir, folderMode); err != nil {
		return fmt.Errorf("making %s: %w", e.dir, err)
	}
	if err := os.WriteFile(e.path(), []byte(entryText(e.app, e.command)), entryMode); err != nil {
		return fmt.Errorf("writing %s: %w", e.path(), err)
	}
	return nil
}

// Disable removes the entry; one that is already gone is not an error. Removal rather than a line
// switching it off, so nothing is left behind for a later version to misread.
func (e Entry) Disable() error {
	if e.problem != nil {
		return e.problem
	}
	if err := os.Remove(e.path()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing %s: %w", e.path(), err)
	}
	return nil
}
