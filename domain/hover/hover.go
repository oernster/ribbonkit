// Package hover decides when an unpinned ribbon opens from its tab and when it collapses back to
// it (FR-615, FR-616). It is told when the pointer arrives and leaves and what the time is; it
// answers whether the ribbon is open and when it next has to be asked again. It reads no clock.
//
// Every operation answers a new State and leaves the receiver as it was.
package hover

import "time"

// Rest is how long the pointer stays on the tab before the ribbon opens (FR-615).
const Rest = 300 * time.Millisecond

// Away is how long the pointer stays off the open ribbon before it collapses (FR-616).
const Away = time.Second

// State is an unpinned ribbon's hover state. The zero value is collapsed, the pointer elsewhere,
// nothing holding it open and nothing pending.
type State struct {
	open, inside, held bool
	// due is when the pending change falls due; zero when none is pending.
	due time.Time
}

// Unpinned answers the state of a ribbon unpinned while shown in full at now: open, collapsing Away
// later unless the pointer is on it (FR-613).
func Unpinned(now time.Time, inside bool) State {
	return State{open: true, inside: inside}.settled(now)
}

// Open answers whether the ribbon is shown in full; false while it is its tab.
func (s State) Open() bool { return s.open }

// Due answers when the state must next be asked through At; false when nothing is pending.
func (s State) Due() (time.Time, bool) { return s.due, !s.due.IsZero() }

// Arrived answers the state once the pointer has come onto the ribbon or its tab at now: a collapsed
// ribbon opens Rest later (FR-615); an open one no longer collapses (FR-616).
func (s State) Arrived(now time.Time) State {
	s.inside = true
	return s.settled(now)
}

// Left answers the state once the pointer has gone off the ribbon or its tab at now: a collapsed
// ribbon no longer opens (FR-615); an open one collapses Away later unless held (FR-616).
func (s State) Left(now time.Time) State {
	s.inside = false
	return s.settled(now)
}

// Held answers the state while something keeps the ribbon as it is (a drag, one of its menus, a
// panel): nothing pending falls due until Released (FR-616).
func (s State) Held() State {
	s.held = true
	s.due = time.Time{}
	return s
}

// Released answers the state once nothing holds the ribbon any longer at now.
func (s State) Released(now time.Time) State {
	s.held = false
	return s.settled(now)
}

// Expanded answers the state of a ribbon shown in full by a command rather than by the pointer, as a
// panel is (FR-616), at now.
func (s State) Expanded(now time.Time) State {
	s.open = true
	return s.settled(now)
}

// At answers the state at now: a pending change that has fallen due is made.
func (s State) At(now time.Time) State {
	if s.due.IsZero() || now.Before(s.due) {
		return s
	}
	s.open = !s.open
	s.due = time.Time{}
	return s
}

// settled answers the state with the change it now awaits: opening Rest from now for a collapsed
// ribbon with the pointer on it; collapsing Away from now for an open one with the pointer off it
// and nothing holding it; else none. A change already pending of the same kind keeps its time.
func (s State) settled(now time.Time) State {
	waits, after := false, time.Duration(0)
	switch {
	case s.held:
	case !s.open && s.inside:
		waits, after = true, Rest
	case s.open && !s.inside:
		waits, after = true, Away
	}
	if !waits {
		s.due = time.Time{}
	} else if s.due.IsZero() {
		s.due = now.Add(after)
	}
	return s
}
