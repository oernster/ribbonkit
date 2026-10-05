// Package atomicfile replaces a file whole, so a failure part way leaves the previous copy as it was
// and a reader never meets half of a new one. Every file a ribbon keeps is written through it: an
// application's settings and forecasts, the kit's occupancy entries.
package atomicfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// tempSuffix ends the name of the temporary file a write goes through before it replaces the real one.
const tempSuffix = ".tmp"

// Write replaces path with data: written to a temporary file in the same folder, flushed to the
// disk, given mode, then renamed over path. On any failure the temporary file is removed and path is
// left as it was. The folder must already exist.
func Write(path string, data []byte, mode os.FileMode) error {
	dir, base := filepath.Split(path)
	temp, err := os.CreateTemp(dir, "."+base+".*"+tempSuffix)
	if err != nil {
		return fmt.Errorf("writing beside %s: %w", path, err)
	}
	name := temp.Name()
	if err := finish(temp, data, mode); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

// finish writes data to file, flushes it to the disk, closes it and gives it mode.
func finish(file *os.File, data []byte, mode os.FileMode) error {
	_, writeErr := file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	return os.Chmod(file.Name(), mode)
}
