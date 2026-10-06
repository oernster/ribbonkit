package platform

import (
	"io"
	"os"
)

// watchShutdown does nothing on macOS, where the application quits when the system asks it to at
// log out, restart or shut down (desktop's ribbon_honour_quit).
func watchShutdown(chan<- os.Signal, io.Writer) {}
