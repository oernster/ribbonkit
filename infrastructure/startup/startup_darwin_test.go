package startup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FR-605.
func TestOpenAtLoginWritesAndRemovesOneAgent(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "LaunchAgents")
	program := "/Applications/TimeRibbon.app/Contents/MacOS/TimeRibbon"
	entry := At(testApp, dir, program)
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
	raw, err := os.ReadFile(filepath.Join(dir, testApp.AppID+".plist"))
	text := string(raw)
	if err != nil || len(names) != 1 || !strings.Contains(text, "<string>"+program+"</string>") ||
		!strings.Contains(text, "<string>"+testApp.AppID+"</string>") {
		t.Errorf("wrote %q among %d agents (%v)", raw, len(names), err)
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

// An agent launchd treats as off is reported off, whoever switched it off.
func TestAnAgentSwitchedOffReadsOff(t *testing.T) {
	t.Parallel()
	entry := At(testApp, t.TempDir(), "/x")
	text := strings.Replace(entryText(testApp, "/x"), "<dict>", "<dict>\n\t<key>Disabled</key>\n\t<true/>", 1)
	if err := os.WriteFile(entry.path(), []byte(text), entryMode); err != nil {
		t.Fatal(err)
	}
	if on, err := entry.Enabled(); err != nil || on {
		t.Errorf("%v %v", on, err)
	}
}

// Absence is off; an agent that is there and cannot be read is a fault.
func TestAnAgentThatCannotBeReadIsAFault(t *testing.T) {
	t.Parallel()
	entry := At(testApp, t.TempDir(), "/x")
	if err := os.Mkdir(entry.path(), folderMode); err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Enabled(); err == nil {
		t.Error("an unreadable agent answered no fault")
	}
}

func TestTheAgentsFolderIsTheOneLaunchdReads(t *testing.T) {
	t.Parallel()
	got, err := agentsDir(func() (string, error) { return "/Users/someone", nil })
	if err != nil || got != "/Users/someone/Library/LaunchAgents" {
		t.Errorf("%q %v", got, err)
	}
	if _, err := agentsDir(func() (string, error) { return "", errors.New("none") }); !errors.Is(err, errNoHome) {
		t.Errorf("no home answered %v", err)
	}
}

// A path is escaped for the property list, so a name holding an ampersand stays one string.
func TestAPathIsEscapedForThePropertyList(t *testing.T) {
	t.Parallel()
	if got := entryText(testApp, "/Apps/A&B <1>"); !strings.Contains(got, "<string>/Apps/A&amp;B &lt;1&gt;</string>") {
		t.Errorf("got %s", got)
	}
}
