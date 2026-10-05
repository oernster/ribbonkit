package platform

import (
	"os"

	"github.com/oernster/ribbonkit/infrastructure/desktop"
)

// Prepare needs to do nothing on Windows, where the tray reads the icon built into the executable and
// a request to end arrives as the window's own close.
func Prepare(*desktop.Desktop, []byte, func(<-chan os.Signal)) {}
