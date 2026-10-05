package occupancy

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/identity"
	"github.com/oernster/ribbonkit/domain/identity/identitytest"
	"github.com/oernster/ribbonkit/domain/placement"
)

// other stands for a second product on the kit, running beside the sample.
var other = identity.App{Name: "OtherRibbon", AppID: "uk.example.OtherRibbon"}

// The rectangles the tests hold: a vertical ribbon at the right edge and its pull out.
var (
	ribbonRect  = placement.Rect{Left: 1744, Top: 418, Right: 1920, Bottom: 614}
	pullOutRect = placement.Rect{Left: 1264, Top: 396, Right: 1744, Bottom: 636}
)

// opened answers the sample's place and another product's in one folder, closed when the test ends.
func opened(t *testing.T, log *strings.Builder) (dir string, sample, beside *Folder) {
	t.Helper()
	dir = t.TempDir()
	sample, beside = Open(dir, identitytest.Sample, log), Open(dir, other, log)
	t.Cleanup(func() {
		sample.Close()
		beside.Close()
	})
	return dir, sample, beside
}

// FR-412: what one running ribbon holds, the other sees; neither sees its own.
func TestARibbonSeesWhatAnotherHolds(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	_, sample, beside := opened(t, &log)
	sample.Hold([]placement.Rect{ribbonRect, pullOutRect})
	if got := beside.Taken(); !slices.Equal(got, []placement.Rect{ribbonRect, pullOutRect}) {
		t.Errorf("the other saw %+v, want the sample's ribbon and pull out", got)
	}
	if got := sample.Taken(); len(got) != 0 {
		t.Errorf("the sample saw itself: %+v", got)
	}
	sample.Hold([]placement.Rect{ribbonRect})
	if got := beside.Taken(); !slices.Equal(got, []placement.Rect{ribbonRect}) {
		t.Errorf("after the pull out closed the other saw %+v", got)
	}
	if log.Len() != 0 {
		t.Errorf("logged %q", log.String())
	}
}

// FR-412: a ribbon that exits takes its entry with it; its lock file stays.
func TestAClosedRibbonIsGone(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	dir, sample, beside := opened(t, &log)
	sample.Hold([]placement.Rect{ribbonRect})
	sample.Close()
	sample.Close()
	if got := beside.Taken(); len(got) != 0 {
		t.Errorf("after the sample closed the other saw %+v", got)
	}
	if _, err := os.Stat(filepath.Join(dir, identitytest.Sample.AppID+entrySuffix)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the entry is still there (%v)", err)
	}
	if _, err := os.Stat(filepath.Join(dir, identitytest.Sample.AppID+lockSuffix)); err != nil {
		t.Errorf("the lock file went: %v", err)
	}
	sample.Hold([]placement.Rect{ribbonRect})
	if got := beside.Taken(); len(got) != 0 {
		t.Errorf("a closed ribbon still held %+v", got)
	}
}

// FR-412: an entry whose lock nobody holds (or which has no lock at all) belongs to a ribbon that has
// gone however it ended; it is passed over and removed.
func TestAGoneRibbonsEntryIsIgnoredAndRemoved(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	dir, sample, _ := opened(t, &log)
	entry := `{"product":"gone","occupied":[{"left":0,"top":0,"right":10,"bottom":10}]}`
	for _, gone := range []string{"uk.example.Crashed", "uk.example.NoLock"} {
		if err := os.WriteFile(filepath.Join(dir, gone+entrySuffix), []byte(entry), fileMode); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "uk.example.Crashed"+lockSuffix), nil, fileMode); err != nil {
		t.Fatal(err)
	}
	if got := sample.Taken(); len(got) != 0 {
		t.Errorf("saw %+v from ribbons that have gone", got)
	}
	for _, gone := range []string{"uk.example.Crashed", "uk.example.NoLock"} {
		if _, err := os.Stat(filepath.Join(dir, gone+entrySuffix)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s's entry is still there (%v)", gone, err)
		}
	}
}

// FR-412, robustness: another ribbon's entry is foreign input. One that is not JSON, too large, holds
// too many rectangles or a rectangle with no area is passed over and said in the log once.
func TestAnEntryThatCannotBeBelievedIsPassedOver(t *testing.T) {
	t.Parallel()
	inverted := `{"occupied":[{"left":10,"top":0,"right":0,"bottom":10}]}`
	many := `{"occupied":[` + strings.TrimSuffix(strings.Repeat(`{"left":0,"top":0,"right":1,"bottom":1},`, maxRects+1), ",") + `]}`
	cases := map[string]struct{ content, reason string }{
		"not JSON":  {"{not json", errBadEntry.Error()},
		"too large": {`{"product":"` + strings.Repeat("x", maxEntryBytes) + `"}`, errEntryTooLarge.Error()},
		"too many":  {many, errBadEntry.Error() + ": 17 of them"},
		"no area":   {inverted, errBadEntry.Error() + ": {Left:10"},
	}
	for name, each := range cases {
		content := each.content
		var log strings.Builder
		dir, sample, beside := opened(t, &log)
		sample.Hold([]placement.Rect{ribbonRect})
		if err := os.WriteFile(filepath.Join(dir, identitytest.Sample.AppID+entrySuffix), []byte(content), fileMode); err != nil {
			t.Fatal(err)
		}
		if got := beside.Taken(); len(got) != 0 {
			t.Errorf("%s: believed %+v", name, got)
		}
		_ = beside.Taken()
		if strings.Count(log.String(), "occupancy: reading uk.example.SampleRibbon's entry") != 1 || !strings.Contains(log.String(), each.reason) {
			t.Errorf("%s: logged %q, want %q once", name, log.String(), each.reason)
		}
	}
}

// FR-412: a second copy of a product cannot hold the entry the first holds; it writes nothing and
// says so, while the first keeps its entry.
func TestASecondCopyHoldsNothing(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	dir, sample, beside := opened(t, &log)
	second := Open(dir, identitytest.Sample, &log)
	sample.Hold([]placement.Rect{ribbonRect})
	second.Hold([]placement.Rect{pullOutRect})
	second.Close()
	if got := beside.Taken(); !slices.Equal(got, []placement.Rect{ribbonRect}) {
		t.Errorf("the other saw %+v, want the first copy's alone", got)
	}
	if !strings.Contains(log.String(), errHeldElsewhere.Error()) {
		t.Errorf("logged %q", log.String())
	}
}

// FR-412: with no folder named (or one that cannot be made) a ribbon is alone; it says why when the
// folder could not be made.
func TestWithoutAFolderARibbonIsAlone(t *testing.T) {
	t.Parallel()
	var silent strings.Builder
	alone := Open("", identitytest.Sample, &silent)
	alone.Hold([]placement.Rect{ribbonRect})
	if got := alone.Taken(); len(got) != 0 || silent.Len() != 0 {
		t.Errorf("no folder: saw %+v, logged %q", got, silent.String())
	}
	alone.Close()
	blocker := filepath.Join(t.TempDir(), "a file")
	if err := os.WriteFile(blocker, nil, fileMode); err != nil {
		t.Fatal(err)
	}
	var log strings.Builder
	blocked := Open(filepath.Join(blocker, kitFolder), identitytest.Sample, &log)
	blocked.Hold([]placement.Rect{ribbonRect})
	if got := blocked.Taken(); len(got) != 0 || !strings.Contains(log.String(), "occupancy: making") {
		t.Errorf("blocked: saw %+v, logged %q", got, log.String())
	}
}

// FR-412: an entry that cannot be written is said in the log; the ribbon carries on. How a write
// fails part way is atomicfile's to test.
func TestAnEntryThatCannotBeWrittenIsSaid(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	dir, _, beside := opened(t, &log)
	if err := os.Mkdir(filepath.Join(dir, other.AppID+entrySuffix), folderMode); err != nil {
		t.Fatal(err)
	}
	beside.Hold([]placement.Rect{ribbonRect})
	if !strings.Contains(log.String(), "occupancy:") || !strings.Contains(log.String(), "this ribbon's entry") {
		t.Errorf("logged %q", log.String())
	}
}

// FR-412: a running ribbon whose entry vanished between the folder being read and the entry being
// opened, as when it exits just then, is no ribbon and no fault.
func TestAnEntryThatVanishesIsNoFault(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	_, sample, _ := opened(t, &log)
	if got := sample.othersRects(other.AppID); len(got) != 0 || log.Len() != 0 {
		t.Errorf("saw %+v, logged %q", got, log.String())
	}
}

// FR-412: a folder that is not there is no other ribbon, said nowhere since nothing went wrong.
func TestAFolderThatIsNotThereIsNoOtherRibbon(t *testing.T) {
	t.Parallel()
	var log strings.Builder
	gone := &Folder{dir: filepath.Join(t.TempDir(), "never made"), id: other.AppID, log: &log}
	if got := gone.Taken(); len(got) != 0 || log.Len() != 0 {
		t.Errorf("saw %+v, logged %q", got, log.String())
	}
}
