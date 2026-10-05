package startup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FR-605.
func TestStartAtSignInWritesAndRemovesOneEntry(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "autostart")
	entry := At(testApp, dir, "/opt/TimeRibbon/timeribbon")
	if on, err := entry.Enabled(); err != nil || on {
		t.Fatalf("before: %v %v", on, err)
	}
	if err := entry.Disable(); err != nil {
		t.Errorf("disabling what is not there: %v", err)
	}
	if err := entry.Enable(); err != nil {
		t.Fatal(err)
	}
	names, _ := os.ReadDir(dir)
	raw, err := os.ReadFile(filepath.Join(dir, testApp.AppID+".desktop"))
	if err != nil || len(names) != 1 || !strings.Contains(string(raw), "\nExec=\"/opt/TimeRibbon/timeribbon\"\n") {
		t.Errorf("wrote %q among %d entries (%v)", raw, len(names), err)
	}
	if on, err := entry.Enabled(); err != nil || !on {
		t.Errorf("after enabling: %v %v", on, err)
	}
	if err := entry.Disable(); err != nil {
		t.Fatal(err)
	}
	if on, _ := entry.Enabled(); on {
		t.Error("still enabled after disabling")
	}
}

// An entry the desktop treats as off is reported off, whoever switched it off.
func TestAnEntrySwitchedOffReadsOff(t *testing.T) {
	t.Parallel()
	for _, line := range []string{hiddenLine, gnomeDisabledLine} {
		dir := t.TempDir()
		entry := At(testApp, dir, "/x")
		if err := os.WriteFile(entry.path(), []byte("[Desktop Entry]\n"+line+"\n"), entryMode); err != nil {
			t.Fatal(err)
		}
		if on, err := entry.Enabled(); err != nil || on {
			t.Errorf("%s: %v %v", line, on, err)
		}
	}
}

// Absence is off; an entry that is there and cannot be read is a fault.
func TestAnEntryThatCannotBeReadIsAFault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	entry := At(testApp, dir, "/x")
	if err := os.Mkdir(entry.path(), folderMode); err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Enabled(); err == nil {
		t.Error("an unreadable entry answered no fault")
	}
}

func TestTheAutostartFolderIsTheOneTheSessionReads(t *testing.T) {
	t.Parallel()
	home := func() (string, error) { return "/home/someone", nil }
	env := func(values map[string]string) func(string) string {
		return func(name string) string { return values[name] }
	}
	cases := []struct {
		name   string
		values map[string]string
		want   string
	}{
		{"default", nil, "/home/someone/.config/autostart"},
		{"XDG_CONFIG_HOME", map[string]string{configHomeVariable: "/cfg"}, "/cfg/autostart"},
		{"Flatpak ignores the private XDG_CONFIG_HOME", map[string]string{configHomeVariable: "/home/someone/.var/app/x/config", flatpakIDVariable: testApp.AppID}, "/home/someone/.config/autostart"},
	}
	for _, c := range cases {
		if got, err := autostartDir(env(c.values), home); err != nil || got != c.want {
			t.Errorf("%s: %q %v, want %q", c.name, got, err, c.want)
		}
	}
	if _, err := autostartDir(env(nil), func() (string, error) { return "", errors.New("none") }); !errors.Is(err, errNoHome) {
		t.Errorf("no home answered %v", err)
	}
}

func TestUnderAFlatpakTheEntryRunsTheFlatpak(t *testing.T) {
	t.Setenv(flatpakIDVariable, testApp.AppID)
	if got := New(testApp, "/app/bin/timeribbon").Command(); got != "flatpak run "+testApp.AppID {
		t.Errorf("command %q", got)
	}
}

// The Desktop Entry specification's reserved characters are escaped inside the quotes, then every
// backslash is doubled as a string value.
func TestAPathIsQuotedForTheExecLine(t *testing.T) {
	t.Parallel()
	if got := execQuoted("/opt/Time Ribbon/a$b"); got != `"/opt/Time Ribbon/a\\$b"` {
		t.Errorf("got %s", got)
	}
}
