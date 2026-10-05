// Package atomicfile replaces a file whole, so a failure part way leaves the previous copy as it was
// and a reader never meets half of a new one. Every file a ribbon keeps is written through it: an
// application's settings and forecasts, the kit's occupancy entries.
package atomicfile

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

// The temporary file a write goes through is named tempPrefix, the real name, tempSeparator, the
// random part os.CreateTemp puts where randomMarker stands, then tempSuffix: ".settings.json.123.tmp".
const (
	tempPrefix    = "."
	tempSeparator = "."
	randomMarker  = "*"
	tempSuffix    = ".tmp"
)

// nameLimit is the longest file name NTFS, APFS and ext4 each allow, in bytes of ASCII.
const nameLimit = 255

// tempPattern answers the os.CreateTemp pattern for a write replacing base.
func tempPattern(base string) string {
	return tempPrefix + base + tempSeparator + randomMarker + tempSuffix
}

// MaxNameLength answers the longest name, in bytes of ASCII, a file Write replaces may have: the
// name limit less what the temporary name adds. os.CreateTemp's random part is a uint32 in decimal,
// so it is at most as long as math.MaxUint32 written out.
func MaxNameLength() int {
	longestRandom := strconv.FormatUint(math.MaxUint32, 10)
	return nameLimit - (len(tempPattern("")) - len(randomMarker) + len(longestRandom))
}

// Write replaces path with data: written to a temporary file in the same folder, flushed to the
// disk, given mode, then renamed over path. On any failure the temporary file is removed and path is
// left as it was. The folder must already exist; a name longer than MaxNameLength is refused by the
// file system.
func Write(path string, data []byte, mode os.FileMode) error {
	dir, base := filepath.Split(path)
	temp, err := os.CreateTemp(dir, tempPattern(base))
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
