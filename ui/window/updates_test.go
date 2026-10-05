package window

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/release"
)

// waitLimit bounds how long a test waits for a check made on a goroutine.
const waitLimit = 5 * time.Second

var offeredUpdate = release.Status{
	Current: "2.0.0", Latest: "v2.1.0", Available: true,
	DownloadURL: "https://example.test/windows", PageURL: "https://example.test/release",
}

func awaitCheck(t *testing.T, checked chan bool) bool {
	t.Helper()
	select {
	case manual := <-checked:
		return manual
	case <-time.After(waitLimit):
		t.Fatal("no update check was made")
		return false
	}
}

// FR-509: an automatic check with nothing new says nothing; one with a newer release shows the
// ribbon as the update panel, carrying the outcome.
func TestAnAutomaticCheckSpeaksOnlyOfANewRelease(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.update = release.Status{Current: "2.0.0", Latest: "v2.0.0"}
	app.checkForUpdate(context.Background(), false)
	if seen.shown != 0 || len(seen.events) != 0 {
		t.Errorf("an up-to-date automatic check showed %d times and sent %v", seen.shown, seen.events)
	}
	service.update = offeredUpdate
	app.checkForUpdate(context.Background(), false)
	opened := slices.IndexFunc(seen.events, func(sent emitted) bool { return sent.event == eventOpenPanel })
	if seen.shown != 1 || opened < 0 || !slices.Equal(seen.events[opened].data, []any{openAtUpdate, updateOf(offeredUpdate)}) {
		t.Errorf("a newer release showed %d times and sent %v", seen.shown, seen.events)
	}
	if !slices.Equal(service.manualChecks, []bool{false, false}) {
		t.Errorf("checks were %v, want two automatic ones", service.manualChecks)
	}
}

// FR-509: a check asked for always answers, GitHub out of reach included; the log says so.
func TestAManualCheckAlwaysAnswers(t *testing.T) {
	t.Parallel()
	app, service, seen, log := newTestApp(t)
	service.update = release.Status{Current: "2.0.0"}
	app.checkForUpdate(context.Background(), true)
	if seen.shown != 1 || !seen.sawEvent(eventOpenPanel, openAtUpdate) || !slices.Equal(service.manualChecks, []bool{true}) {
		t.Errorf("showed %d times, sent %v after checks %v", seen.shown, seen.events, service.manualChecks)
	}
	if !strings.Contains(log.String(), "could not reach GitHub") {
		t.Errorf("the log says %q", log.String())
	}
}

// Help's item runs a manual check.
func TestCheckForUpdatesInHelpAsksForACheck(t *testing.T) {
	t.Parallel()
	app, service, _, _ := newTestApp(t)
	service.checked = make(chan bool, 1)
	app.act(menus.Updates)
	if !awaitCheck(t, service.checked) {
		t.Error("Help's check was made as an automatic one")
	}
}

// A panic in a check is logged and ends nothing.
func TestAPanicInACheckIsLogged(t *testing.T) {
	t.Parallel()
	app, service, seen, log := newTestApp(t)
	service.panicOnUpdate = true
	app.checkForUpdate(context.Background(), true)
	if !strings.Contains(log.String(), "recovered from planted panic while checking for an update") || seen.shown != 0 {
		t.Errorf("logged %q, showed %d times", log.String(), seen.shown)
	}
}

// FR-509: the watch checks after its delay and then at each interval, all automatically; it stops
// when the run ends.
func TestTheWatchChecksAfterTheStartThenAtEachIntervalUntilTheEnd(t *testing.T) {
	t.Parallel()
	app, service, _, _ := newTestApp(t)
	service.checked = make(chan bool)
	app.updates.delay, app.updates.every = time.Millisecond, time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() {
		app.watchForUpdates(ctx)
		close(stopped)
	}()
	for range 2 {
		if awaitCheck(t, service.checked) {
			t.Error("the watch made a manual check")
		}
	}
	cancel()
	for {
		select {
		case <-service.checked:
		case <-stopped:
			return
		case <-time.After(waitLimit):
			t.Fatal("the watch did not stop when the run ended")
		}
	}
}

func TestAWatchEndedBeforeItsFirstCheckChecksNothing(t *testing.T) {
	t.Parallel()
	app, service, _, _ := newTestApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	app.watchForUpdates(ctx)
	if len(service.manualChecks) != 0 {
		t.Errorf("checks %v after the run ended", service.manualChecks)
	}
}

// FR-509: Download opens this platform's download, else the release page; before an offer it opens
// nothing. A browser that cannot open says where the download is.
func TestDownloadOpensWhatWasOffered(t *testing.T) {
	t.Parallel()
	app, _, seen, _ := newTestApp(t)
	if err := app.OpenUpdate(); !errors.Is(err, errNothingOffered) {
		t.Errorf("before an offer Download answered %v", err)
	}
	app.updates.offer(offeredUpdate)
	pageOnly := offeredUpdate
	pageOnly.DownloadURL = ""
	if err := app.OpenUpdate(); err != nil {
		t.Fatal(err)
	}
	app.updates.offer(pageOnly)
	if err := app.OpenUpdate(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(seen.browsed, []string{offeredUpdate.DownloadURL, offeredUpdate.PageURL}) {
		t.Errorf("browsed %v", seen.browsed)
	}
	seen.browseErr = errPlanted
	if err := app.OpenUpdate(); !errors.Is(err, errPlanted) || !strings.Contains(err.Error(), offeredUpdate.PageURL) {
		t.Errorf("a browser that would not open answered %v", err)
	}
}

// FR-509: Skip keeps the offered version; before an offer it keeps nothing.
func TestSkipKeepsTheOfferedVersion(t *testing.T) {
	t.Parallel()
	app, service, _, _ := newTestApp(t)
	if err := app.SkipUpdate(); !errors.Is(err, errNothingOffered) {
		t.Errorf("before an offer Skip answered %v", err)
	}
	app.updates.offer(offeredUpdate)
	if err := app.SkipUpdate(); err != nil || !slices.Equal(service.skipped, []string{offeredUpdate.Latest}) {
		t.Errorf("Skip answered %v and kept %v", err, service.skipped)
	}
}
