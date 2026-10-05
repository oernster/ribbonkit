//go:build windows || darwin

package desktop

import (
	"fmt"
	"sync"
	"time"
)

// pointerEvery is how often Windows and macOS read where the pointer is while an unpinned ribbon is
// shown (Oliver, 2026-09-28). A pass shorter than this can be missed, which is harmless since one that
// short must not open the ribbon (FR-615); a departure is still read at the next reading.
const pointerEvery = 50 * time.Millisecond

// pointerTracker holds the channel that stops the reading; nil while the pointer is not read.
type pointerTracker struct {
	guard sync.Mutex
	stop  chan struct{}
}

// TrackPointer starts or stops reporting the pointer arriving on the ribbon and leaving it (FR-615,
// FR-616). Where the pointer is is read every pointerEvery: on macOS the page hears nothing while
// the ribbon is not the active application; the position read in Go was measured to see every
// arrival and departure on both (REQUIREMENTS section 2.3). The page is not used on either. Starting reports where the pointer is.
func (d *Desktop) TrackPointer(ribbon Window, on bool) {
	d.pointer.guard.Lock()
	defer d.pointer.guard.Unlock()
	if !on {
		if d.pointer.stop != nil {
			close(d.pointer.stop)
			d.pointer.stop = nil
		}
		return
	}
	if d.pointer.stop != nil || ribbon == 0 {
		return
	}
	stop := make(chan struct{})
	d.pointer.stop = stop
	go d.readPointer(ribbon, stop)
}

// readPointer reports each change of the pointer being on ribbon until stop closes. A failure to read
// is logged once until a reading succeeds again; a panic is logged, which ends the reading.
func (d *Desktop) readPointer(ribbon Window, stop <-chan struct{}) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(d.log, "desktop: recovered from %v while reading the pointer; the ribbon opens from its tab again once it is next shown\n", failure)
		}
	}()
	ticker := time.NewTicker(pointerEvery)
	defer ticker.Stop()
	known, inside, failing := false, false, false
	for {
		now, err := pointerInside(ribbon)
		switch {
		case err != nil && !failing:
			fmt.Fprintf(d.log, "desktop: reading the pointer: %v\n", err)
			failing = true
		case err == nil:
			failing = false
			if !known || now != inside {
				known, inside = true, now
				d.send(pointerEvent(now))
			}
		}
		select {
		case <-stop:
			return
		case <-ticker.C:
		}
	}
}
