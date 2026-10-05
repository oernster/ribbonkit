package runlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTheLogIsMadeAndEachRunAddsItsStartLine(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), testApp.Name)
	for _, at := range []time.Time{time.Date(2026, 9, 27, 21, 37, 0, 0, time.UTC), time.Date(2026, 9, 28, 6, 37, 0, 0, time.UTC)} {
		log, err := Open(dir, testApp, at)
		if err != nil {
			t.Fatal(err)
		}
		_ = log.Close()
	}
	raw, _ := os.ReadFile(filepath.Join(dir, FileName(testApp)))
	want := testApp.Name + " started 2026-09-27 21:37:00\n" + testApp.Name + " started 2026-09-28 06:37:00\n"
	if string(raw) != want {
		t.Errorf("got %q", raw)
	}
}

func TestALogOverTheLimitIsStartedAfresh(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName(testApp)), []byte(strings.Repeat("x", MaxBytes+1)), fileMode); err != nil {
		t.Fatal(err)
	}
	log, err := Open(dir, testApp, time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	_ = log.Close()
	raw, _ := os.ReadFile(filepath.Join(dir, FileName(testApp)))
	if string(raw) != testApp.Name+" started 2026-09-27 00:00:00\n" {
		t.Errorf("got %d bytes", len(raw))
	}
}

func TestAFolderThatCannotBeMadeIsAnswered(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, fileMode); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(filepath.Join(blocker, testApp.Name), testApp, time.Now()); err == nil {
		t.Error("a log under a file opened")
	}
}

// NFR-O-1: what is written to standard error after Keep reaches the log. Run in a child process so
// the test binary's own standard error is left alone.
func TestLogReceivesStandardError(t *testing.T) {
	if dir := os.Getenv("RIBBONKIT_RUNLOG_CHILD"); dir != "" {
		log, err := Open(dir, testApp, time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))
		if err != nil {
			os.Exit(2)
		}
		if err := Keep(log); err != nil {
			os.Exit(3)
		}
		_, _ = os.Stderr.WriteString("reported\n")
		os.Exit(0)
	}
	dir := t.TempDir()
	child := testCommand(t, dir)
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("child: %v %s", err, out)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, FileName(testApp)))
	if !strings.HasSuffix(string(raw), "reported\n") {
		t.Errorf("log holds %q", raw)
	}
}
