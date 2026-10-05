//go:build windows

package setup

// Which conversation this run is having (FR-801), decided once from one reading of the machine so
// the screen, its heading, its boxes and its buttons cannot drift apart. The route never becomes
// Uninstall: removal is a screen reachable from every route, so the one behind it stays the one to
// come back to. Nothing here touches the machine, so every state is a test.

// Route names the screen setup opens on.
type Route string

const (
	// RouteInstall is for a machine with nothing recorded.
	RouteInstall Route = "install"
	// RouteUpdate is for a machine holding an older version than setup carries.
	RouteUpdate Route = "update"
	// RouteDowngrade is for a machine holding a newer version than setup carries: Go back.
	RouteDowngrade Route = "downgrade"
	// RouteManage is for a machine holding this very version: Repair, Reinstall or Uninstall.
	RouteManage Route = "manage"
)

// Choices are the three boxes an install applies (FR-805).
type Choices struct {
	StartMenu        bool
	Desktop          bool
	StartWithWindows bool
}

// Existing is what the machine already holds, read once.
type Existing struct {
	// Installed says the Apps list records the application; Version is the version it records.
	Installed bool
	Version   string
	// Choices are the boxes as they stand on the machine: the shortcuts present and the Start with
	// Windows value present.
	Choices Choices
}

// freshChoices are the boxes on a machine with nothing installed (FR-805): the Start Menu ticked;
// the Desktop shortcut and Start with Windows unticked.
var freshChoices = Choices{StartMenu: true}

// RouteFor decides the route from what the machine holds and the version setup carries.
func RouteFor(existing Existing, carried string) Route {
	if !existing.Installed {
		return RouteInstall
	}
	switch Compare(carried, existing.Version) {
	case Newer:
		return RouteUpdate
	case Older:
		return RouteDowngrade
	}
	return RouteManage
}

// Offered answers the boxes a screen opens on: what the machine already holds where the application is
// installed, so a shortcut someone declined is never offered back as though they had asked for
// it; the fresh defaults where it is not.
func Offered(existing Existing) Choices {
	if !existing.Installed {
		return freshChoices
	}
	return existing.Choices
}
