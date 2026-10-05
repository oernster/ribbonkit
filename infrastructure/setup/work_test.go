//go:build windows

package setup

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// lines is a Recorder keeping what it is told.
type lines struct{ kept []string }

func (l *lines) Record(line string) { l.kept = append(l.kept, line) }

// The bar moves by each step's measured weight rather than by the count of steps; it names the step.
func TestTheBarIsWeightedByMeasuredTime(t *testing.T) {
	t.Parallel()
	var seen []Progress
	steps := []Step{
		{Name: "slow", Weight: 3, Do: func() error { return nil }},
		{Name: "quick", Weight: 1, Do: func() error { return nil }},
	}
	log := &lines{}
	if err := Run(steps, log, func(p Progress) { seen = append(seen, p) }); err != nil {
		t.Fatal(err)
	}
	want := []Progress{{0, "slow"}, {75, "quick"}, {percentWhole, finishedStep}}
	if !reflect.DeepEqual(seen, want) {
		t.Errorf("reported %v, want %v", seen, want)
	}
	if strings.Join(log.kept, "|") != "start: slow|done: slow|start: quick|done: quick" {
		t.Errorf("logged %v", log.kept)
	}
}

// The first failure stops the work, names its step and is logged; nothing after it runs.
func TestAFailingStepStopsTheWorkAndSaysWhy(t *testing.T) {
	t.Parallel()
	refused := errors.New("the disk is full")
	ran := false
	steps := []Step{
		{Name: "Writing the files", Weight: 1, Do: func() error { return refused }},
		{Name: "after", Weight: 1, Do: func() error { ran = true; return nil }},
	}
	log := &lines{}
	err := Run(steps, log, func(Progress) {})
	if !errors.Is(err, refused) || err.Error() != "Writing the files: the disk is full" {
		t.Errorf("got %v", err)
	}
	if ran {
		t.Error("a step ran after the failure")
	}
	if last := log.kept[len(log.kept)-1]; last != "failed: Writing the files: the disk is full" {
		t.Errorf("logged %q last", last)
	}
	var none []Progress
	if err := Run(nil, log, func(p Progress) { none = append(none, p) }); err != nil || none[0].Percent != percentWhole {
		t.Errorf("no work reported %v (%v)", none, err)
	}
}

// Each line reaches the file as it is written, before the log is closed.
func TestTheStepLogIsOnDiskAsItGoes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "setup.log")
	log, err := OpenStepLog(path)
	if err != nil || log.Path() != path {
		t.Fatalf("opening: %v, path %q", err, log.Path())
	}
	log.Record("start: Writing the files")
	raw, err := os.ReadFile(path)
	if err != nil || !strings.HasSuffix(string(raw), " start: Writing the files\n") {
		t.Errorf("the file holds %q (%v) before it is closed", raw, err)
	}
	if err := log.Close(); err != nil {
		t.Error(err)
	}
	log.Record("after closing, written nowhere")
	if again, _ := os.ReadFile(path); string(again) != string(raw) {
		t.Errorf("a line reached the file after closing: %q", again)
	}
}

// A log that cannot be opened still answers a log, one that writes nowhere, so setup runs on.
func TestALogThatCannotBeOpenedWritesNowhere(t *testing.T) {
	t.Parallel()
	log, err := OpenStepLog(filepath.Join(t.TempDir(), "absent", "setup.log"))
	if err == nil || log.Path() != "" {
		t.Fatalf("got %v, path %q", err, log.Path())
	}
	log.Record("nowhere")
	if err := log.Close(); err != nil {
		t.Error(err)
	}
}

// A write that fails switches the log off rather than failing setup.
func TestALogWhoseWriteFailsFallsSilent(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "setup.log")
	log, err := OpenStepLog(path)
	if err != nil {
		t.Fatal(err)
	}
	log.sync = func() error { return errors.New("the disk went away") }
	log.Record("first")
	log.Record("second")
	_ = log.Close()
	if raw, _ := os.ReadFile(path); strings.Contains(string(raw), "second") {
		t.Errorf("the log went on writing after a failure: %q", raw)
	}
}
