//go:build windows

package setup

import (
	"fmt"

	"github.com/oernster/ribbonkit/domain/identity"
	"github.com/oernster/ribbonkit/infrastructure/startup"
)

// StartupEntry is the Start with Windows value: startup.Entry, the one Settings writes (FR-805).
type StartupEntry interface {
	Enabled() (bool, error)
	Enable() error
	Disable() error
}

// StartWithWindows answers app's Start with Windows value for a program. It is the value Settings
// writes through the same package, so setup's box and the application's switch cannot disagree.
func StartWithWindows(app identity.App) func(program string) StartupEntry {
	return func(program string) StartupEntry { return startup.New(app, program) }
}

// Carried is what setup brings to an install: the payload, the version it carries and setup's own
// path, which is copied in as the uninstaller.
type Carried struct {
	Payload string
	Self    string
	Version string
}

// Machine is the user's machine as setup acts on it: the places it writes, the Apps list entry,
// the Start with Windows value and the removal of the install folder once setup has closed.
type Machine struct {
	product         Product
	places          Places
	record          Record
	entryFor        func(program string) StartupEntry
	deleteAfterExit func(dir string) error
}

// NewMachine answers the machine product is installed on, over its parts. The setup program passes
// the real ones; a test passes temporary folders, a scratch registry key and a stand-in for the
// removal.
func NewMachine(product Product, places Places, record Record, entryFor func(string) StartupEntry, deleteAfterExit func(string) error) Machine {
	return Machine{product: product, places: places, record: record, entryFor: entryFor, deleteAfterExit: deleteAfterExit}
}

// Places answers the folders this machine writes.
func (m Machine) Places() Places { return m.places }

// Read answers what the machine already holds, once, so every screen opens on what is true.
func (m Machine) Read() (Existing, error) {
	version, installed := m.record.Version()
	startWithWindows, err := m.entryFor(m.places.Program()).Enabled()
	if err != nil {
		return Existing{}, fmt.Errorf("reading Start with Windows: %w", err)
	}
	return Existing{
		Installed: installed,
		Version:   version,
		Choices: Choices{
			StartMenu:        shortcutPresent(m.places.StartMenu, m.product.shortcut()),
			Desktop:          shortcutPresent(m.places.Desktop, m.product.shortcut()),
			StartWithWindows: startWithWindows,
		},
	}, nil
}

// InstallSteps are Install, Update, Go back and Reinstall alike (FR-802): the files, the
// uninstaller beside them, the Apps list entry, then the boxes as chosen.
func (m Machine) InstallSteps(carried Carried, choices Choices) []Step {
	steps := []Step{
		{Name: "Writing the files", Weight: weightExtract, Do: func() error {
			return ExtractZip(carried.Payload, m.places.InstallDir)
		}},
		{Name: "Leaving the uninstaller beside them", Weight: weightUninstaller, Do: func() error {
			return CopyFile(carried.Self, m.places.Uninstaller())
		}},
		{Name: "Adding " + m.product.App.Name + " to the Apps list", Weight: weightRecord, Do: func() error {
			return m.register(carried.Version)
		}},
	}
	return append(steps, m.choiceSteps(choices)...)
}

// RepairSteps write the files again as an install does, keeping the shortcuts and the Start with
// Windows value exactly as they are on the machine (FR-804).
func (m Machine) RepairSteps(carried Carried) ([]Step, error) {
	existing, err := m.Read()
	if err != nil {
		return nil, err
	}
	return m.InstallSteps(carried, existing.Choices), nil
}

// ChoiceSteps apply the boxes alone, as a box on the Installed screen does the moment it changes.
func (m Machine) ChoiceSteps(choices Choices) []Step { return m.choiceSteps(choices) }

// UninstallSteps remove the shortcuts, the Start with Windows value and the Apps list entry, then
// leave the install folder to be deleted once setup has closed (FR-806). Forgetting the settings
// deletes the settings folder and nothing else.
func (m Machine) UninstallSteps(forget bool) []Step {
	program := m.places.Program()
	steps := []Step{
		m.shortcutStep(startMenuPlace, m.places.StartMenu, program, false),
		m.shortcutStep(desktopPlace, m.places.Desktop, program, false),
		{Name: "Removing Start with Windows", Weight: weightStartup, Do: m.entryFor(program).Disable},
		{Name: "Removing " + m.product.App.Name + " from the Apps list", Weight: weightRecord, Do: m.record.Remove},
	}
	if forget {
		steps = append(steps, Step{Name: "Forgetting your settings", Weight: weightForget, Do: func() error {
			return RemoveTree(m.places.Settings)
		}})
	}
	return append(steps, Step{Name: "Removing the files once setup closes", Weight: weightSchedule, Do: func() error {
		return m.deleteAfterExit(m.places.InstallDir)
	}})
}

// The places a shortcut goes, as the steps name them.
const (
	startMenuPlace = "Start Menu"
	desktopPlace   = "Desktop"
)

// choiceSteps place or clear each shortcut, then write or delete the Start with Windows value.
func (m Machine) choiceSteps(choices Choices) []Step {
	program := m.places.Program()
	entry := m.entryFor(program)
	startWithWindows := Step{Name: "Setting Start with Windows", Weight: weightStartup, Do: entry.Disable}
	if choices.StartWithWindows {
		startWithWindows.Do = entry.Enable
	}
	return []Step{
		m.shortcutStep(startMenuPlace, m.places.StartMenu, program, choices.StartMenu),
		m.shortcutStep(desktopPlace, m.places.Desktop, program, choices.Desktop),
		startWithWindows,
	}
}

// shortcutStep places the shortcut in folder where it is wanted and clears it where not. Clearing is
// weighted apart because it takes no shell object and was measured far quicker.
func (m Machine) shortcutStep(place, folder, program string, wanted bool) Step {
	step := Step{Name: "Clearing the " + place + " shortcut", Weight: weightClearShortcut}
	if wanted {
		step = Step{Name: "Placing the " + place + " shortcut", Weight: weightShortcut}
	}
	link := m.product.shortcut()
	step.Do = func() error { return placeShortcut(folder, link, program, wanted) }
	return step
}

// register writes the Apps list entry for the files now in place.
func (m Machine) register(version string) error {
	size, err := DirSizeKB(m.places.InstallDir)
	if err != nil {
		return err
	}
	return m.record.Write(UninstallInfo{
		Version:      version,
		InstallDir:   m.places.InstallDir,
		UninstallExe: m.places.Uninstaller(),
		IconPath:     m.places.Program(),
		EstimatedKB:  size,
	})
}
