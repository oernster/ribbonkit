package startup

import (
	"errors"
	"fmt"

	"github.com/oernster/ribbonkit/domain/identity"

	"golang.org/x/sys/windows/registry"
)

// RunKey is the per-user key Windows starts programs from at sign-in.
const RunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// Entry is the application's StartupEntry port over one value.
type Entry struct {
	key     string
	value   string
	program string
}

// New answers the entry for program, the full path of the application's executable, under RunKey.
func New(app identity.App, program string) Entry {
	return At(app, RunKey, program)
}

// At answers the entry for program under key, a path beneath HKCU. Only a test names a key other
// than RunKey: a scratch key of its own, so the real value is never touched.
func At(app identity.App, key, program string) Entry {
	return Entry{key: key, value: app.Name, program: program}
}

// Command answers what the value holds: the quoted program path and no arguments, so a sign-in
// start shows the ribbon as a normal launch does (FR-605).
func (e Entry) Command() string {
	return `"` + e.program + `"`
}

// Enabled answers whether the value is present.
func (e Entry) Enabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, e.key, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("opening %s: %w", e.key, err)
	}
	defer key.Close()
	_, _, err = key.GetStringValue(e.value)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", e.value, err)
	}
	return true, nil
}

// Enable writes the value.
func (e Entry) Enable() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, e.key, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening %s: %w", e.key, err)
	}
	defer key.Close()
	if err := key.SetStringValue(e.value, e.Command()); err != nil {
		return fmt.Errorf("writing %s: %w", e.value, err)
	}
	return nil
}

// Disable deletes the value; one that is already gone is not an error.
func (e Entry) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, e.key, registry.SET_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening %s: %w", e.key, err)
	}
	defer key.Close()
	if err := key.DeleteValue(e.value); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("deleting %s: %w", e.value, err)
	}
	return nil
}
