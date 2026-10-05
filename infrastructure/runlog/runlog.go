// Package runlog keeps the log a run leaves (NFR-O-1): a line naming when the run started, then
// whatever the run reports, in the application's log (TimeRibbon.log for TimeRibbon) inside the settings folder.
//
// A windowed program's standard error reaches nobody: on Windows it is handed a handle of zero, so
// everything written there is lost, the Go runtime's own panic report included. Keep points the
// run's error output and os.Stderr at the log as the first act of the run, so a crash leaves a
// record rather than a silence (ported from Bridge Talk, where it was measured against Go 1.26.3).
// Each platform supplies Keep in a file of its own.
package runlog

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/oernster/ribbonkit/domain/identity"
)

// FileName names app's log inside the settings folder.
func FileName(app identity.App) string { return app.Name + ".log" }

// MaxBytes is the size past which a run starts the log afresh, so it cannot grow without end.
const MaxBytes = 1 << 20

const (
	startedLayout             = "2006-01-02 15:04:05"
	folderMode    fs.FileMode = 0o700
	fileMode      fs.FileMode = 0o600
)

// Open opens the log in dir for a run started at started, making dir where it is missing, then
// adds the start line. A log over MaxBytes is started afresh.
func Open(dir string, app identity.App, started time.Time) (*os.File, error) {
	if err := os.MkdirAll(dir, folderMode); err != nil {
		return nil, fmt.Errorf("making %s: %w", dir, err)
	}
	path := filepath.Join(dir, FileName(app))
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if info, err := os.Stat(path); err == nil && info.Size() > MaxBytes {
		flags |= os.O_TRUNC
	}
	log, err := os.OpenFile(path, flags, fileMode)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	if _, err := fmt.Fprintf(log, "%s started %s\n", app.Name, started.Format(startedLayout)); err != nil {
		_ = log.Close()
		return nil, fmt.Errorf("writing to %s: %w", path, err)
	}
	return log, nil
}
