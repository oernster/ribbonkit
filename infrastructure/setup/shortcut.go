//go:build windows

package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

// shellProgramID names the Windows shell automation object that reads and writes shortcuts.
const shellProgramID = "WScript.Shell"

// The shortcut properties setup writes and a test reads back.
const (
	shortcutTarget     = "TargetPath"
	shortcutIcon       = "IconLocation"
	shortcutWorkingDir = "WorkingDirectory"
)

// shortcutPresent reports whether folder holds the shortcut saved as name.
func shortcutPresent(folder, name string) bool {
	_, err := os.Stat(filepath.Join(folder, name))
	return err == nil
}

// placeShortcut puts a shortcut to program in folder, saved as name, when wanted and takes it away
// when not, so unticking a box on a reinstall removes the shortcut rather than leaving a stale one.
func placeShortcut(folder, name, program string, wanted bool) error {
	link := filepath.Join(folder, name)
	if !wanted {
		if err := os.Remove(link); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("removing %s: %w", link, err)
		}
		return nil
	}
	if err := os.MkdirAll(folder, dirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", folder, err)
	}
	return createShortcut(link, program, filepath.Dir(program))
}

// withShell runs fn against the shell automation object with COM started, then releases what it
// took. COM belongs to a thread, so the goroutine stays on one for the call.
func withShell(fn func(shell *ole.IDispatch) error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stop, err := startCOM()
	if err != nil {
		return fmt.Errorf("starting COM: %w", err)
	}
	defer stop()
	unknown, err := oleutil.CreateObject(shellProgramID)
	if err != nil {
		return fmt.Errorf("creating %s: %w", shellProgramID, err)
	}
	defer unknown.Release()
	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("reaching %s: %w", shellProgramID, err)
	}
	defer shell.Release()
	return fn(shell)
}

// startCOM starts COM on this thread, answering with what undoes it. A thread already running COM
// under the same model answers S_FALSE, which still needs its uninitialise; one running under the
// other model answers RPC_E_CHANGED_MODE, usable but not started here, so not stopped here.
func startCOM() (func(), error) {
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err == nil {
		return ole.CoUninitialize, nil
	}
	var failure *ole.OleError
	if !errors.As(err, &failure) {
		return nil, err
	}
	switch windows.Handle(failure.Code()) {
	case windows.S_FALSE:
		return ole.CoUninitialize, nil
	case windows.RPC_E_CHANGED_MODE:
		return func() {}, nil
	}
	return nil, err
}

// openShortcut answers the shortcut object for path: the one on disk, else a new one to fill in.
func openShortcut(shell *ole.IDispatch, path string) (*ole.IDispatch, error) {
	result, err := oleutil.CallMethod(shell, "CreateShortcut", path)
	if err != nil {
		return nil, fmt.Errorf("opening the shortcut %s: %w", path, err)
	}
	return result.ToIDispatch(), nil
}

// createShortcut writes a .lnk through the shell object, handing each path over as a value, so no
// character in a path is ever read as anything but itself.
func createShortcut(linkPath, target, workDir string) error {
	return withShell(func(shell *ole.IDispatch) error {
		link, err := openShortcut(shell, linkPath)
		if err != nil {
			return err
		}
		defer link.Release()
		properties := []struct{ name, value string }{
			{shortcutTarget, target}, {shortcutIcon, target}, {shortcutWorkingDir, workDir},
		}
		for _, property := range properties {
			if _, err := oleutil.PutProperty(link, property.name, property.value); err != nil {
				return fmt.Errorf("setting %s on the shortcut %s: %w", property.name, linkPath, err)
			}
		}
		if _, err := oleutil.CallMethod(link, "Save"); err != nil {
			return fmt.Errorf("saving the shortcut %s: %w", linkPath, err)
		}
		return nil
	})
}
