// Package occupancy is the folder every ribbon on the kit shares with the others running for the
// same user, of whatever product, so that none lands on another (FR-412). Each ribbon keeps an entry
// there of the rectangles it occupies and holds a lock on a file beside it while it runs; an entry
// whose lock nobody holds belongs to a ribbon that has gone, however it ended. A Folder is the
// arranger's Neighbours.
//
// The lock is a file of its own because a locked range on Windows cannot be read by others. Lock
// files are never removed: a ribbon starting up may have opened its lock but not yet taken it. A
// lock file left behind costs one empty file per product.
package occupancy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oernster/ribbonkit/domain/identity"
	"github.com/oernster/ribbonkit/domain/placement"
)

const (
	// kitFolder names the folder every product on the kit shares.
	kitFolder   = "ribbonkit"
	entrySuffix = ".json"
	lockSuffix  = ".lock"
	tempSuffix  = ".tmp"
	// maxEntryBytes caps what is read of another ribbon's entry; a real one, two rectangles, is well
	// under a kilobyte.
	maxEntryBytes = 64 << 10
	// maxRects caps the rectangles one entry may claim; a ribbon and its pull out are two.
	maxRects = 16
	// lockAttempts and lockRetry bound the wait for this ribbon's own lock, which another ribbon
	// finding out whether the entry is stale holds for an instant.
	lockAttempts = 20
	lockRetry    = 25 * time.Millisecond
	folderMode   = 0o700
	fileMode     = 0o600
)

// ErrNoFolder is answered when the environment names no folder for the kit's shared state.
var ErrNoFolder = errors.New("the environment names no folder for the ribbons to share")

// errEntryTooLarge and errBadEntry describe another ribbon's entry this one will not believe.
var (
	errEntryTooLarge = errors.New("larger than any ribbon writes")
	errBadEntry      = errors.New("not rectangles a ribbon occupies")
	errHeldElsewhere = errors.New("another copy of this ribbon holds its entry, so this one keeps none")
)

// Dir answers the folder the ribbons share, reading the environment through lookup.
func Dir(lookup func(string) (string, bool)) (string, error) {
	return dir(lookup)
}

// entryFile is an entry as written: the product it is and the rectangles it occupies.
type entryFile struct {
	Product  string    `json:"product"`
	Occupied []rectDTO `json:"occupied"`
}

// rectDTO is a rectangle in physical pixels.
type rectDTO struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

// Folder is this ribbon's place in the shared folder. It is safe to call from several goroutines.
type Folder struct {
	dir string
	id  string
	log io.Writer

	mutex sync.Mutex
	// lock is this ribbon's lock file while it holds it; nil while it holds none, when it writes no
	// entry but still keeps off the others.
	lock *os.File
	// reported is the last failure written to the log, so one that repeats at every placement is
	// written once.
	reported string
}

// Open takes this ribbon's place, app, in the folder dir, writing any failure to log. It never
// fails: a ribbon that cannot take its place keeps off the others it can read and is otherwise alone.
// An empty dir, where the environment named none, is a ribbon alone.
func Open(dir string, app identity.App, log io.Writer) *Folder {
	f := &Folder{dir: dir, id: app.AppID, log: log}
	if dir == "" {
		return f
	}
	if err := os.MkdirAll(dir, folderMode); err != nil {
		f.report(fmt.Errorf("making %s: %w", dir, err))
		return f
	}
	lock, err := os.OpenFile(f.path(f.id, lockSuffix), os.O_CREATE|os.O_RDWR, fileMode)
	if err != nil {
		f.report(fmt.Errorf("opening this ribbon's lock: %w", err))
		return f
	}
	for range lockAttempts {
		held, err := tryLock(lock)
		if err != nil {
			f.report(fmt.Errorf("locking this ribbon's entry: %w", err))
			break
		}
		if held {
			f.lock = lock
			return f
		}
		time.Sleep(lockRetry)
	}
	if f.reported == "" {
		f.report(errHeldElsewhere)
	}
	_ = lock.Close()
	return f
}

// Taken answers the rectangles the other running ribbons occupy. An entry nobody holds is removed;
// one that cannot be read is passed over and said in the log.
func (f *Folder) Taken() []placement.Rect {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.dir == "" {
		return nil
	}
	found, err := os.ReadDir(f.dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			f.report(fmt.Errorf("reading %s: %w", f.dir, err))
		}
		return nil
	}
	var taken []placement.Rect
	for _, each := range found {
		id, isEntry := strings.CutSuffix(each.Name(), entrySuffix)
		if !isEntry || id == f.id {
			continue
		}
		taken = append(taken, f.othersRects(id)...)
	}
	return taken
}

// othersRects answers what the ribbon id occupies while it runs; nothing once it has gone, when its
// entry is removed.
func (f *Folder) othersRects(id string) []placement.Rect {
	lock, err := os.OpenFile(f.path(id, lockSuffix), os.O_RDWR, fileMode)
	if errors.Is(err, fs.ErrNotExist) {
		f.removeStale(id)
		return nil
	}
	if err != nil {
		f.report(fmt.Errorf("opening %s's lock: %w", id, err))
		return nil
	}
	defer lock.Close()
	free, err := tryLock(lock)
	if err != nil {
		f.report(fmt.Errorf("testing %s's lock: %w", id, err))
		return nil
	}
	if free {
		_ = unlock(lock)
		f.removeStale(id)
		return nil
	}
	rects, err := readEntry(f.path(id, entrySuffix))
	if errors.Is(err, fs.ErrNotExist) {
		// It exited between the folder being read and its entry being opened.
		return nil
	}
	if err != nil {
		f.report(fmt.Errorf("reading %s's entry: %w", id, err))
		return nil
	}
	return rects
}

// removeStale removes the entry of a ribbon that has gone.
func (f *Folder) removeStale(id string) {
	if err := os.Remove(f.path(id, entrySuffix)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		f.report(fmt.Errorf("removing %s's stale entry: %w", id, err))
	}
}

// readEntry reads an entry another ribbon wrote, believing nothing in it unchecked.
func readEntry(path string) ([]placement.Rect, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxEntryBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxEntryBytes {
		return nil, errEntryTooLarge
	}
	var entry entryFile
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("%w: %w", errBadEntry, err)
	}
	if len(entry.Occupied) > maxRects {
		return nil, fmt.Errorf("%w: %d of them", errBadEntry, len(entry.Occupied))
	}
	rects := make([]placement.Rect, 0, len(entry.Occupied))
	for _, each := range entry.Occupied {
		if each.Right <= each.Left || each.Bottom <= each.Top {
			return nil, fmt.Errorf("%w: %+v", errBadEntry, each)
		}
		rects = append(rects, placement.Rect{Left: each.Left, Top: each.Top, Right: each.Right, Bottom: each.Bottom})
	}
	return rects, nil
}

// Hold replaces this ribbon's entry with occupied, written whole before it replaces the last one so
// a reader never meets half of it. A ribbon holding no lock writes nothing.
func (f *Folder) Hold(occupied []placement.Rect) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.lock == nil {
		return
	}
	entry := entryFile{Product: f.id, Occupied: make([]rectDTO, 0, len(occupied))}
	for _, each := range occupied {
		entry.Occupied = append(entry.Occupied, rectDTO{Left: each.Left, Top: each.Top, Right: each.Right, Bottom: each.Bottom})
	}
	// A struct of a string and whole numbers always marshals.
	data, _ := json.Marshal(entry)
	final := f.path(f.id, entrySuffix)
	temporary := final + tempSuffix
	if err := os.WriteFile(temporary, data, fileMode); err != nil {
		f.report(fmt.Errorf("writing this ribbon's entry: %w", err))
		return
	}
	if err := os.Rename(temporary, final); err != nil {
		f.report(fmt.Errorf("replacing this ribbon's entry: %w", err))
	}
}

// Close removes this ribbon's entry and lets its lock go, as the application exits. A ribbon that
// ends any other way leaves an entry nobody holds, which the next ribbon to look removes.
func (f *Folder) Close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.lock == nil {
		return
	}
	if err := os.Remove(f.path(f.id, entrySuffix)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		f.report(fmt.Errorf("removing this ribbon's entry: %w", err))
	}
	_ = unlock(f.lock)
	_ = f.lock.Close()
	f.lock = nil
}

// path answers the file id's entry or lock is kept in.
func (f *Folder) path(id, suffix string) string {
	return filepath.Join(f.dir, id+suffix)
}

// report writes err to the log unless it is the failure written last.
func (f *Folder) report(err error) {
	text := err.Error()
	if text == f.reported {
		return
	}
	f.reported = text
	fmt.Fprintf(f.log, "occupancy: %s\n", text)
}
