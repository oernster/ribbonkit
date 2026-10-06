package desktop

import "testing"

// Wails refuses every request from macOS to quit and hands it to a close handler that hides the
// ribbon, so a log out or restart was interrupted. Once the ribbon is set up the application agrees.
func TestTheApplicationQuitsWhenMacOSAsks(t *testing.T) {
	withRefusingDelegate(func() {
		if quits() {
			t.Fatal("the stand-in for Wails' delegate already agrees to quit, so the test proves nothing")
		}
		if err := HideFromTaskbar(0); err != nil {
			t.Fatal(err)
		}
		if !quits() {
			t.Error("asked to quit by macOS, the application refused")
		}
	})
}
