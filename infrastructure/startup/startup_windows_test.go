package startup

import (
	"crypto/rand"
	"testing"

	"golang.org/x/sys/windows/registry"
)

// scratch answers an entry under a key of its own beneath HKCU\Software, deleted when the test
// ends. The real Run key is never touched.
func scratch(t *testing.T) Entry {
	t.Helper()
	parent := `Software\TimeRibbonTest`
	key := parent + `\` + rand.Text()
	t.Cleanup(func() {
		_ = registry.DeleteKey(registry.CURRENT_USER, key)
		_ = registry.DeleteKey(registry.CURRENT_USER, parent)
	})
	return At(testApp, key, `C:\Users\Someone\AppData\Local\Programs\TimeRibbon\TimeRibbon.exe`)
}

// FR-605.
func TestStartWithWindowsWritesAndRemovesOneValue(t *testing.T) {
	entry := scratch(t)
	if on, err := entry.Enabled(); err != nil || on {
		t.Fatalf("before: %v %v", on, err)
	}
	if err := entry.Disable(); err != nil {
		t.Errorf("disabling what is not there: %v", err)
	}
	if err := entry.Enable(); err != nil {
		t.Fatal(err)
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, entry.key, registry.QUERY_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	command, _, err := key.GetStringValue(testApp.Name)
	names, _ := key.ReadValueNames(0)
	key.Close()
	if err != nil || command != `"C:\Users\Someone\AppData\Local\Programs\TimeRibbon\TimeRibbon.exe"` || len(names) != 1 {
		t.Errorf("wrote %q among %v (%v)", command, names, err)
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
	if err := entry.Disable(); err != nil {
		t.Errorf("disabling twice: %v", err)
	}
}

func TestTheRealEntryNamesTheRunKey(t *testing.T) {
	t.Parallel()
	entry := New(testApp, `C:\x\TimeRibbon.exe`)
	if entry.key != RunKey || entry.value != testApp.Name || entry.Command() != `"C:\x\TimeRibbon.exe"` {
		t.Errorf("got %+v", entry)
	}
}
