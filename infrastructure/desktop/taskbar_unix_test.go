//go:build linux || darwin

package desktop

import "testing"

// FR-101: the ribbon stays off the taskbar. On Linux the window is marked to stay off the taskbar
// and the workspace switcher; on macOS the application becomes an accessory, with no Dock icon.
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
