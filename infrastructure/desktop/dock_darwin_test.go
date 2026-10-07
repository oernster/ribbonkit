package desktop

import "testing"

// Wails makes the application regular as it finishes launching, which made the Dock record the
// ribbon as a recent app at every launch. Wrapped as the kit wraps Wails' delegate, the switch is
// refused and the application stays an accessory; setting the policy afterwards still works.
func TestWailsCannotMakeTheRibbonRegularAsItLaunches(t *testing.T) {
	if staysAccessoryThroughLaunch() {
		t.Fatal("the stand-in for Wails' delegate already leaves the application an accessory, so the test proves nothing")
	}
	wrapRegularDelegate()
	if !staysAccessoryThroughLaunch() {
		t.Error("finishing launching made the application regular, which puts it in the Dock's recents")
	}
	if !policySettable() {
		t.Error("after launching, the application could no longer be made regular: AppKit's method was not restored")
	}
}
