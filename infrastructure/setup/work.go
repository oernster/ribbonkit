//go:build windows

package setup

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// Step is one piece of setup's work: named for the reader, weighted by the time it was measured to
// take, so the bar moves as the work does rather than by counting steps.
type Step struct {
	Name   string
	Weight int
	Do     func() error
}

// Progress is how far the work has got: the share of the measured time already spent, in percent,
// and the step now running.
type Progress struct {
	Percent int
	Step    string
}

// Recorder keeps the step log. Each line is flushed as it is written, since the failures that
// matter most are the ones that never raise.
type Recorder interface {
	Record(line string)
}

const (
	// percentWhole is the bar at the end of the work.
	percentWhole = 100
	// finishedStep is what the bar says once every step has run.
	finishedStep = "Done."
)

// Run performs the steps in order, reporting progress before each and recording each start and
// each outcome. The first step to fail stops the work and answers why, naming the step.
func Run(steps []Step, log Recorder, report func(Progress)) error {
	total := 0
	for _, step := range steps {
		total += step.Weight
	}
	spent := 0
	for _, step := range steps {
		report(Progress{Percent: share(spent, total), Step: step.Name})
		log.Record("start: " + step.Name)
		if err := step.Do(); err != nil {
			log.Record("failed: " + step.Name + ": " + err.Error())
			return fmt.Errorf("%s: %w", step.Name, err)
		}
		log.Record("done: " + step.Name)
		spent += step.Weight
	}
	report(Progress{Percent: percentWhole, Step: finishedStep})
	return nil
}

// share answers spent as a percentage of total; no work at all is already whole.
func share(spent, total int) int {
	if total == 0 {
		return percentWhole
	}
	return spent * percentWhole / total
}

// StepLog is the record of what setup did, one line per event, synced to disk as it is written.
type StepLog struct {
	out   io.Writer
	sync  func() error
	close func() error
	path  string
}

// logFlags open the step log to add to what an earlier run left.
const logFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY

// logPerm lets the user read and write the step log.
const logPerm = 0o644

// stampLayout is the time each line begins with.
const stampLayout = time.RFC3339

// OpenStepLog opens the step log at path. Where it cannot be opened the log answered writes
// nowhere, so setup still runs; the error says why.
func OpenStepLog(path string) (*StepLog, error) {
	file, err := os.OpenFile(path, logFlags, logPerm)
	if err != nil {
		return &StepLog{out: io.Discard, sync: nothing, close: nothing}, fmt.Errorf("opening the step log %s: %w", path, err)
	}
	return &StepLog{out: file, sync: file.Sync, close: file.Close, path: path}, nil
}

// nothing stands in for the sync and the close of a log that writes nowhere.
func nothing() error { return nil }

// Path answers where the log is written; empty where it could not be opened.
func (l *StepLog) Path() string { return l.path }

// Close closes the log file.
func (l *StepLog) Close() error {
	l.out, l.sync = io.Discard, nothing
	return l.close()
}

// Record writes one line with the time it happened and flushes it to disk.
func (l *StepLog) Record(line string) {
	_, err := fmt.Fprintf(l.out, "%s %s\n", time.Now().Format(stampLayout), line)
	err = errors.Join(err, l.sync())
	if err != nil {
		l.out, l.sync = io.Discard, nothing
	}
}
