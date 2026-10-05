//go:build darwin || linux

package occupancy

import (
	"errors"
	"os"
	"syscall"
)

// tryLock takes file's lock without waiting; false where another holder has it. flock's lock belongs
// to the open file, so it holds across processes and across two Flatpak sandboxes sharing the folder.
func tryLock(file *os.File) (bool, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}
	return err == nil, err
}

// unlock lets file's lock go.
func unlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
