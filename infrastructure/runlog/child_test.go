package runlog

import (
	"os"
	"os/exec"
	"testing"
)

// testCommand answers this test binary run again as the child of TestLogReceivesStandardError.
func testCommand(t *testing.T, dir string) *exec.Cmd {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^TestLogReceivesStandardError$")
	command.Env = append(os.Environ(), "RIBBONKIT_RUNLOG_CHILD="+dir)
	return command
}
