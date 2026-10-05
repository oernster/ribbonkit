//go:build windows

package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// The install folder holds the running setup program when the Apps list opens it, so the folder can
// only go once setup has closed (FR-806). A hidden PowerShell waits on setup's process id, which cmd
// cannot do, then deletes. The folder reaches it as an environment value rather than typed into the
// script, so no character in a path is read as anything but itself.

const (
	// deletionShell is the program that waits for setup to close and then deletes.
	deletionShell = "powershell.exe"
	// deletionDirVariable carries the folder to delete in that program's environment.
	deletionDirVariable = "RIBBONKIT_SETUP_DELETE_DIR"
)

// dirDeletion answers the arguments and the environment entry that make deletionShell wait for pid
// to exit and then delete dir. A process already gone is not waited for, so a late start still runs.
func dirDeletion(pid int, dir string) (args, env []string) {
	script := "Wait-Process -Id " + strconv.Itoa(pid) + " -ErrorAction SilentlyContinue; " +
		"Remove-Item -LiteralPath $env:" + deletionDirVariable + " -Recurse -Force -ErrorAction SilentlyContinue"
	return []string{"-NoProfile", "-NonInteractive", "-Command", script}, []string{deletionDirVariable + "=" + dir}
}

// DeleteAfterExit starts the hidden deletion of dir, to run once this process has exited.
func DeleteAfterExit(dir string) error { return deleteAfter(os.Getpid(), dir) }

// deleteAfter starts the deletion of dir waiting on pid, so a test can play setup with a process it
// is able to close. It is started beside dir: a process holds its working folder, so one started
// inside dir would empty it and leave it standing.
func deleteAfter(pid int, dir string) error {
	args, env := dirDeletion(pid, dir)
	cmd := exec.Command(deletionShell, args...)
	cmd.Dir = filepath.Dir(dir)
	cmd.Env = append(os.Environ(), env...)
	cmd.SysProcAttr = hidden()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting the removal of %s: %w", dir, err)
	}
	return cmd.Process.Release()
}
