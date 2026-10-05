//go:build linux || darwin

package runlog

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// Keep points the run's error output at log: descriptor 2, which the runtime writes its reports
// through, then os.Stderr.
func Keep(log *os.File) error {
	if err := unix.Dup2(int(log.Fd()), unix.Stderr); err != nil {
		return fmt.Errorf("sending error output to %s: %w", log.Name(), err)
	}
	os.Stderr = log
	return nil
}
