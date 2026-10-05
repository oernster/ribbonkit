//go:build windows

package setup

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// uninstallKeys is where the per-user Apps list keeps its entries, one key per application.
const uninstallKeys = `Software\Microsoft\Windows\CurrentVersion\Uninstall\`

// The Apps list entry's value names.
const (
	valueDisplayName     = "DisplayName"
	valueDisplayVersion  = "DisplayVersion"
	valueInstallLocation = "InstallLocation"
	valueUninstallString = "UninstallString"
	valueModifyPath      = "ModifyPath"
	valueDisplayIcon     = "DisplayIcon"
	valuePublisher       = "Publisher"
	valueNoModify        = "NoModify"
	valueNoRepair        = "NoRepair"
	valueEstimatedSize   = "EstimatedSize"
)

// offered is the NoModify and NoRepair value that leaves Windows offering the button (FR-802).
const offered = 0

// UninstallInfo is what the Apps list entry records.
type UninstallInfo struct {
	Version      string
	InstallDir   string
	UninstallExe string
	IconPath     string
	EstimatedKB  uint32
}

// Record is the Apps list entry: one key under HKCU, naming the product and its publisher.
type Record struct {
	key       string
	name      string
	publisher string
}

// AppsList answers product's entry in the per-user Apps list.
func AppsList(product Product) Record {
	return Record{key: uninstallKeys + product.App.Name, name: product.App.Name, publisher: product.Publisher}
}

// values is the text the entry holds, by value name. The uninstaller path is quoted the way
// Windows reads a command line; Go's %q escapes the separators and would name nowhere.
func (r Record) values(info UninstallInfo) map[string]string {
	quoted := `"` + info.UninstallExe + `"`
	return map[string]string{
		valueDisplayName:     r.name,
		valueDisplayVersion:  info.Version,
		valueInstallLocation: info.InstallDir,
		valueUninstallString: quoted + " " + UninstallFlag,
		valueModifyPath:      quoted,
		valueDisplayIcon:     info.IconPath,
		valuePublisher:       r.publisher,
	}
}

// Write records the application, with Modify and Repair offered: both reopen the uninstaller copy,
// which then opens on the Installed screen (FR-802).
func (r Record) Write(info UninstallInfo) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, r.key, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening %s: %w", r.key, err)
	}
	defer key.Close()
	for name, value := range r.values(info) {
		if err := key.SetStringValue(name, value); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	numbers := map[string]uint32{valueNoModify: offered, valueNoRepair: offered, valueEstimatedSize: info.EstimatedKB}
	for name, value := range numbers {
		if err := key.SetDWordValue(name, value); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	return nil
}

// Remove deletes the entry; one already gone is not an error.
func (r Record) Remove() error {
	err := registry.DeleteKey(registry.CURRENT_USER, r.key)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("deleting %s: %w", r.key, err)
	}
	return nil
}

// Version answers the version the entry records and whether there is an entry at all.
func (r Record) Version() (string, bool) {
	key, err := registry.OpenKey(registry.CURRENT_USER, r.key, registry.QUERY_VALUE)
	if err != nil {
		return "", false
	}
	defer key.Close()
	version, _, err := key.GetStringValue(valueDisplayVersion)
	if err != nil {
		return "", false
	}
	return version, true
}
