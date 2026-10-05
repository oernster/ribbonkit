package desktop

import "testing"

// FR-101: on Linux the ribbon is marked to stay off the taskbar and the workspace switcher. macOS
// keeps its Dock icon (Amendment 36), so this holds for Linux alone.
func TestTheRibbonIsKeptOffTheTaskbar(t *testing.T) {
	ribbon := newTestWindow(testApp.Name)
	defer closeTestWindow(ribbon)
	if err := HideFromTaskbar(ribbon); err != nil {
		t.Fatal(err)
	}
	if skips, err := skipsTaskbar(ribbon); err != nil || !skips {
		t.Errorf("skips the taskbar %v (%v)", skips, err)
	}
}
