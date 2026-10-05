package occupancy

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// lockedBytes is the length of the range locked: one byte of a file that holds nothing.
const lockedBytes = 1

// tryLock takes file's lock without waiting; false where another holder has it.
func tryLock(file *os.File) (bool, error) {
	var overlapped windows.Overlapped
	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, lockedBytes, 0, &overlapped)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return false, nil
	}
	return err == nil, err
}

// unlock lets file's lock go.
func unlock(file *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, lockedBytes, 0, &overlapped)
}
