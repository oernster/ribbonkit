//go:build windows

package setup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows/registry"

	"github.com/oernster/ribbonkit/infrastructure/startup"
)

// carried is a payload holding a program and the licence, with a stand-in for setup itself.
func carried(t *testing.T) Carried {
	t.Helper()
	self := filepath.Join(t.TempDir(), "SampleRibbonSetup.exe")
	write(t, self, "the setup program")
	return Carried{Payload: zipOf(t, map[string]string{sample.Exe(): "the program", LicenceFile: "terms"}), Self: self, Version: "0.1.0"}
}

// runAll runs steps with nowhere to report, failing the test on the first refusal.
func runAll(t *testing.T, steps []Step) {
	t.Helper()
	if err := Run(steps, &lines{}, func(Progress) {}); err != nil {
		t.Fatal(err)
	}
}

// shortcutTargetOf reads a saved shortcut's target back through the shell object that wrote it.
func shortcutTargetOf(t *testing.T, path string) string {
	t.Helper()
	var target string
	err := withShell(func(shell *ole.IDispatch) error {
		link, err := openShortcut(shell, path)
		if err != nil {
			return err
		}
		defer link.Release()
		value, err := oleutil.GetProperty(link, shortcutTarget)
		if err != nil {
			return err
		}
		target = value.ToString()
		return value.Clear()
	})
	if err != nil {
		t.Fatal(err)
	}
	return target
}

// recorded reads one text value of the record's key.
func recorded(t *testing.T, record Record, name string) string {
	t.Helper()
	key, err := registry.OpenKey(registry.CURRENT_USER, record.key, registry.QUERY_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	defer key.Close()
	value, _, err := key.GetStringValue(name)
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return value
}

// FR-802: files, the uninstaller beside them, the Apps list entry with Modify and Repair, then the
// boxes as chosen.
func TestAnInstallWritesEveryPartInPlace(t *testing.T) {
	f := newFixture(t)
	runAll(t, f.machine.InstallSteps(carried(t), Choices{StartMenu: true, StartWithWindows: true}))

	if got, _ := os.ReadFile(f.places.Program()); string(got) != "the program" {
		t.Errorf("the program holds %q", got)
	}
	if got, _ := os.ReadFile(f.places.Uninstaller()); string(got) != "the setup program" {
		t.Errorf("the uninstaller holds %q", got)
	}
	if version, installed := f.record.Version(); !installed || version != "0.1.0" {
		t.Errorf("the Apps list records %q (%v)", version, installed)
	}
	if recorded(t, f.record, valueInstallLocation) != f.places.InstallDir {
		t.Error("the Apps list entry names another folder")
	}
	key, _ := registry.OpenKey(registry.CURRENT_USER, f.record.key, registry.QUERY_VALUE)
	noModify, _, modifyErr := key.GetIntegerValue(valueNoModify)
	noRepair, _, repairErr := key.GetIntegerValue(valueNoRepair)
	key.Close()
	if modifyErr != nil || repairErr != nil || noModify != offered || noRepair != offered {
		t.Errorf("Modify and Repair are not offered: %d %d (%v %v)", noModify, noRepair, modifyErr, repairErr)
	}
	if got := shortcutTargetOf(t, filepath.Join(f.places.StartMenu, sample.shortcut())); got != f.places.Program() {
		t.Errorf("the Start Menu shortcut targets %q", got)
	}
	if shortcutPresent(f.places.Desktop, sample.shortcut()) {
		t.Error("a Desktop shortcut was placed with its box unticked")
	}
	if on, err := startup.At(sample.App, f.startupKey, f.places.Program()).Enabled(); err != nil || !on {
		t.Errorf("Start with Windows is %v (%v)", on, err)
	}
}

// FR-802: the entry's commands name the uninstaller in plain quotes; Go's %q would escape the
// separators and name nowhere.
func TestTheUninstallEntryNamesTheRealPath(t *testing.T) {
	t.Parallel()
	path := filepath.Join(`C:\Users\Someone\AppData\Local`, programsSubdir, sample.App.Name, UninstallExeName)
	values := AppsList(sample).values(UninstallInfo{UninstallExe: path, Version: "0.1.0"})
	want := map[string]string{
		valueUninstallString: `"` + path + `" ` + UninstallFlag,
		valueModifyPath:      `"` + path + `"`,
		valueDisplayName:     sample.App.Name,
		valueDisplayVersion:  "0.1.0",
		valuePublisher:       sample.Publisher,
	}
	for name, expected := range want {
		if values[name] != expected {
			t.Errorf("%s = %s, want %s", name, values[name], expected)
		}
	}
}

// FR-804: the boxes read what is on the machine; Repair keeps them exactly so.
func TestTheBoxesReflectWhatIsOnTheMachine(t *testing.T) {
	f := newFixture(t)
	existing, err := f.machine.Read()
	if err != nil || existing.Installed || existing.Choices != (Choices{}) {
		t.Fatalf("an empty machine read %+v (%v)", existing, err)
	}
	runAll(t, f.machine.InstallSteps(carried(t), Choices{Desktop: true, StartWithWindows: true}))
	want := Choices{Desktop: true, StartWithWindows: true}
	existing, err = f.machine.Read()
	if err != nil || !existing.Installed || existing.Version != "0.1.0" || existing.Choices != want {
		t.Fatalf("after installing read %+v (%v), want %+v", existing, err, want)
	}

	// A damaged install: the program gone. Repair puts it back and keeps every box as it was.
	if err := os.Remove(f.places.Program()); err != nil {
		t.Fatal(err)
	}
	steps, err := f.machine.RepairSteps(carried(t))
	if err != nil {
		t.Fatal(err)
	}
	runAll(t, steps)
	if !exists(f.places.Program()) {
		t.Error("repair did not put the program back")
	}
	if after, _ := f.machine.Read(); after.Choices != want {
		t.Errorf("repair changed the boxes to %+v, want %+v", after.Choices, want)
	}
}

// FR-805: setup's box writes the very value Settings writes, through the same package.
func TestStartWithWindowsIsTheSameValueSettingsWrites(t *testing.T) {
	t.Parallel()
	program := filepath.Join(`C:\Users\Someone\AppData\Local`, programsSubdir, sample.App.Name, sample.Exe())
	entry, ok := StartWithWindows(sample.App)(program).(startup.Entry)
	if !ok || entry != startup.New(sample.App, program) {
		t.Fatalf("setup's entry is %#v, want the one Settings writes, %#v", StartWithWindows(sample.App)(program), startup.New(sample.App, program))
	}
	if entry.Command() != `"`+program+`"` {
		t.Errorf("the value holds %s", entry.Command())
	}
}

// FR-806: uninstall removes the shortcuts, the value and the entry, leaves the folder to go once
// setup closes and keeps the settings; forgetting deletes the settings folder and nothing else.
func TestForgettingRemovesOnlyTheSettingsFolder(t *testing.T) {
	for _, forget := range []bool{false, true} {
		f := newFixture(t)
		runAll(t, f.machine.InstallSteps(carried(t), Choices{StartMenu: true, Desktop: true, StartWithWindows: true}))
		settingsFile := filepath.Join(f.places.Settings, "settings.json")
		neighbour := filepath.Join(filepath.Dir(f.places.Settings), "SomeoneElse", "keep.txt")
		write(t, settingsFile, "{}")
		write(t, neighbour, "not ours")

		runAll(t, f.machine.UninstallSteps(forget))

		if shortcutPresent(f.places.StartMenu, sample.shortcut()) || shortcutPresent(f.places.Desktop, sample.shortcut()) {
			t.Errorf("forget %v: a shortcut survived", forget)
		}
		if on, _ := startup.At(sample.App, f.startupKey, f.places.Program()).Enabled(); on {
			t.Errorf("forget %v: Start with Windows survived", forget)
		}
		if _, installed := f.record.Version(); installed {
			t.Errorf("forget %v: the Apps list entry survived", forget)
		}
		if len(f.scheduled) != 1 || f.scheduled[0] != f.places.InstallDir {
			t.Errorf("forget %v: scheduled %v for removal", forget, f.scheduled)
		}
		if exists(settingsFile) == forget {
			t.Errorf("forget %v: the settings file there is %v", forget, exists(settingsFile))
		}
		if !exists(neighbour) || !exists(f.places.Program()) {
			t.Errorf("forget %v: something beside the settings folder went", forget)
		}
	}
}

// On the Installed screen each box applies at once: ticking places, unticking clears.
func TestTheBoxesApplyAtOnce(t *testing.T) {
	f := newFixture(t)
	runAll(t, f.machine.InstallSteps(carried(t), Choices{StartMenu: true}))
	runAll(t, f.machine.ChoiceSteps(Choices{Desktop: true, StartWithWindows: true}))
	if got, _ := f.machine.Read(); got.Choices != (Choices{Desktop: true, StartWithWindows: true}) {
		t.Errorf("after applying, the machine reads %+v", got.Choices)
	}
}

// A machine whose Start with Windows value cannot be read says so rather than guessing.
func TestAReadingThatFailsIsReported(t *testing.T) {
	t.Parallel()
	refused := errors.New("refused")
	broken := NewMachine(sample, Places{}, Record{key: scratchKey(t)}, func(string) StartupEntry { return failingEntry{refused} }, nil)
	if _, err := broken.Read(); !errors.Is(err, refused) {
		t.Errorf("got %v", err)
	}
	if _, err := broken.RepairSteps(Carried{}); !errors.Is(err, refused) {
		t.Errorf("repair got %v", err)
	}
}

// failingEntry refuses everything with err.
type failingEntry struct{ err error }

func (f failingEntry) Enabled() (bool, error) { return false, f.err }
func (f failingEntry) Enable() error          { return f.err }
func (f failingEntry) Disable() error         { return f.err }
