package heldfile

import (
	"fmt"
	"syscall"
)

// noSharing lets nobody else open, write or delete the file while it is held.
const noSharing = 0

// Hold opens the file at path with no sharing and answers the function that lets it go. Until then
// every other open of the file is refused; so is its removal.
func Hold(path string) (release func(), err error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("holding %q: %w", path, err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, noSharing, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, fmt.Errorf("holding %s: %w", path, err)
	}
	return func() { _ = syscall.CloseHandle(handle) }, nil
}
