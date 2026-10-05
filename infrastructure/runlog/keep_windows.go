package runlog

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// Keep points the run's error output at log: the handle the runtime writes its reports through,
// then os.Stderr.
func Keep(log *os.File) error {
	if err := windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(log.Fd())); err != nil {
		return fmt.Errorf("sending error output to %s: %w", log.Name(), err)
	}
	os.Stderr = log
	return nil
}
