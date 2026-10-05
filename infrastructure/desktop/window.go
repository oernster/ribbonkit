// Package desktop is a ribbon's integration with the desktop below the window (CON-7): the
// notification area icon and its menu (FR-501 to FR-503), the end of a move (FR-404), the display,
// time and resume broadcasts (FR-209, FR-406) and the operations on the ribbon's own window (FR-101,
// FR-401, FR-405). Each platform supplies them in files of its own; this one holds what they share.
package desktop

import (
	"errors"
	"time"

	"github.com/oernster/ribbonkit/application/shell"
)

// Window is the shell port's, named here for the same reason as Event (events.go).
type Window = shell.Window

// findAttempts and findPause bound the wait for the ribbon's window to exist after Wails starts.
const (
	findAttempts = 50
	findPause    = 20 * time.Millisecond
)

// ErrRibbonNotFound is answered when the ribbon's window does not appear.
var ErrRibbonNotFound = errors.New("the ribbon's window was not found")
