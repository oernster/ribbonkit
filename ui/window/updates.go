package window

// The update check's place in the window's life (FR-509): once shortly after the start, then once a
// day while it runs, plus whenever Help asks. An automatic check says nothing unless a release the
// user has not skipped is newer; a manual one always answers. The outcome opens the update panel on
// the ribbon, shown for it. The page never names an address: it asks Go to open what Go offered.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oernster/ribbonkit/application/release"
)

// When the automatic check runs: after the start has settled, then daily (FR-509).
const (
	updateCheckDelay = 3 * time.Second
	updateCheckEvery = 24 * time.Hour
)

// errNothingOffered refuses Download or Skip before any update check has offered a release.
var errNothingOffered = errors.New("no update has been offered")

// updateWatch is the automatic check's timing and the outcome last put before the user.
type updateWatch struct {
	delay time.Duration
	every time.Duration

	mutex   sync.Mutex
	offered release.Status
}

func (w *updateWatch) offer(status release.Status) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.offered = status
}

func (w *updateWatch) current() release.Status {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.offered
}

// watchForUpdates checks after the start has settled, then once a day, until ctx ends.
func (a *Window) watchForUpdates(ctx context.Context) {
	first := time.NewTimer(a.updates.delay)
	defer first.Stop()
	select {
	case <-ctx.Done():
		return
	case <-first.C:
	}
	a.checkForUpdate(ctx, false)
	daily := time.NewTicker(a.updates.every)
	defer daily.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-daily.C:
			a.checkForUpdate(ctx, false)
		}
	}
}

// checkForUpdate runs one check. Where it has something to say, it shows the ribbon as the update
// panel. It runs on a goroutine of its own, so a panic is logged here rather than ending the run.
func (a *Window) checkForUpdate(ctx context.Context, manual bool) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(a.log, "recovered from %v while checking for an update\n", failure)
		}
	}()
	status := a.service.CheckForUpdate(ctx, manual)
	if status.Latest == "" {
		fmt.Fprintln(a.log, "the update check could not reach GitHub")
	}
	if !manual && !status.Available {
		return
	}
	a.updates.offer(status)
	a.show()
	a.emit(eventOpenPanel, openAtUpdate, updateOf(status))
}

// OpenUpdate hands the offered release to the browser: this platform's download, else the release
// page (FR-509). Where the browser cannot be opened, the refusal gives the address.
func (a *Window) OpenUpdate() error {
	status := a.updates.current()
	address := status.DownloadURL
	if address == "" {
		address = status.PageURL
	}
	if address == "" {
		return errNothingOffered
	}
	if err := a.browse(address); err != nil {
		return fmt.Errorf("your browser could not be opened on the download (%w); it is at %s", err, address)
	}
	return nil
}

// SkipUpdate keeps the offered release as one the automatic check never offers again (FR-509).
func (a *Window) SkipUpdate() error {
	status := a.updates.current()
	if status.Latest == "" {
		return errNothingOffered
	}
	return a.service.SkipUpdate(status.Latest)
}
