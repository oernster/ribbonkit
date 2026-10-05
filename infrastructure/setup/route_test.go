//go:build windows

package setup

import "testing"

// FR-801: major, minor then patch as numbers, anything after a hyphen ignored, a missing or
// non-numeric field counted as zero.
func TestCompareOrdersVersions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b string
		want Relation
	}{
		{"0.1.0", "0.1.0", Same},
		{"0.2.0", "0.1.9", Newer},
		{"0.1.9", "0.2.0", Older},
		{"1.0.0", "0.99.99", Newer},
		{"0.1.10", "0.1.9", Newer},
		{"0.1.0-beta", "0.1.0", Same},
		{"0.1", "0.1.0", Same},
		{"0.x.1", "0.0.1", Same},
		{"", "0.0.0", Same},
		{" 2.0.0 ", "1.9.9", Newer},
		{"0.0.0-dev", "0.1.0", Older},
	}
	for _, each := range cases {
		if got := Compare(each.a, each.b); got != each.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", each.a, each.b, got, each.want)
		}
	}
}

// FR-801: one reading decides the route; the route never becomes Uninstall.
func TestTheRouteFollowsTheReading(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		existing Existing
		want     Route
	}{
		{"nothing recorded", Existing{}, RouteInstall},
		{"an older version recorded", Existing{Installed: true, Version: "0.0.9"}, RouteUpdate},
		{"a newer version recorded", Existing{Installed: true, Version: "0.2.0"}, RouteDowngrade},
		{"the same version recorded", Existing{Installed: true, Version: "0.1.0"}, RouteManage},
	}
	for _, each := range cases {
		if got := RouteFor(each.existing, "0.1.0"); got != each.want {
			t.Errorf("%s: got %s, want %s", each.name, got, each.want)
		}
	}
}

// FR-805 and the house rule: a fresh machine opens on the defaults; an installed one opens on what
// it already holds, so a declined shortcut is never offered back.
func TestTheBoxesOpenOnWhatIsTrue(t *testing.T) {
	t.Parallel()
	fresh := Offered(Existing{Choices: Choices{Desktop: true, StartWithWindows: true}})
	if fresh != (Choices{StartMenu: true}) {
		t.Errorf("a fresh machine offered %+v", fresh)
	}
	held := Choices{Desktop: true, StartWithWindows: true}
	if got := Offered(Existing{Installed: true, Version: "0.1.0", Choices: held}); got != held {
		t.Errorf("an installed machine offered %+v, want %+v", got, held)
	}
}
