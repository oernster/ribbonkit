package hover

import (
	"testing"
	"time"
)

var start = time.Date(2026, 9, 28, 20, 0, 0, 0, time.UTC)

func at(offset time.Duration) time.Time { return start.Add(offset) }

// opened answers a ribbon the pointer opened by resting on its tab from start, the pointer still on it.
func opened(t *testing.T) State {
	t.Helper()
	s := State{}.Arrived(start).At(at(Rest))
	if !s.Open() {
		t.Fatal("resting on the tab did not open the ribbon")
	}
	return s
}

func wantDue(t *testing.T, s State, want time.Time) {
	t.Helper()
	if got, pending := s.Due(); !pending || !got.Equal(want) {
		t.Errorf("due %v (pending %v), want %v", got, pending, want)
	}
}

func wantNothingDue(t *testing.T, s State) {
	t.Helper()
	if got, pending := s.Due(); pending {
		t.Errorf("due %v, want nothing pending", got)
	}
}

func TestACollapsedRibbonWaitsForThePointer(t *testing.T) {
	t.Parallel()
	s := State{}
	if s.Open() {
		t.Error("the zero state is open")
	}
	wantNothingDue(t, s)
}

// FR-615: the ribbon opens once the pointer has rested on the tab for Rest, not before.
func TestTheRibbonOpensAfterTheRest(t *testing.T) {
	t.Parallel()
	s := State{}.Arrived(start)
	wantDue(t, s, at(Rest))
	if s.At(at(Rest - time.Millisecond)).Open() {
		t.Error("opened before the rest was over")
	}
	s = s.At(at(Rest))
	if !s.Open() {
		t.Error("did not open once the rest was over")
	}
	wantNothingDue(t, s)
	// A second arrival while resting does not restart the rest.
	wantDue(t, State{}.Arrived(start).Arrived(at(Rest/2)), at(Rest))
}

// FR-615: a pointer crossing the tab on its way elsewhere does not open it (a 0.18 s pass on macOS).
func TestAPassingPointerDoesNotOpenIt(t *testing.T) {
	t.Parallel()
	s := State{}.Arrived(start).Left(at(180 * time.Millisecond))
	wantNothingDue(t, s)
	if s.At(at(time.Minute)).Open() {
		t.Error("a passing pointer opened the ribbon")
	}
}

// FR-616: the open ribbon collapses once the pointer has been off it for Away, not before.
func TestTheRibbonCollapsesASecondAfterThePointerLeaves(t *testing.T) {
	t.Parallel()
	left := at(5 * time.Second)
	s := opened(t).Left(left)
	wantDue(t, s, left.Add(Away))
	if !s.At(left.Add(Away - time.Millisecond)).Open() {
		t.Error("collapsed before the second was over")
	}
	s = s.At(left.Add(Away))
	if s.Open() {
		t.Error("did not collapse once the second was over")
	}
	wantNothingDue(t, s)
}

// FR-616: the pointer coming back within the second keeps the ribbon open; the next departure
// counts afresh. No hand run exercised this; it is held here.
func TestAReturningPointerKeepsItOpen(t *testing.T) {
	t.Parallel()
	s := opened(t).Left(at(5 * time.Second)).Arrived(at(5*time.Second + 500*time.Millisecond))
	wantNothingDue(t, s)
	if !s.At(at(time.Minute)).Open() {
		t.Error("a pointer that came back did not keep the ribbon open")
	}
	later := at(10 * time.Second)
	wantDue(t, s.Left(later), later.Add(Away))
}

// Growing the window under the pointer raises a false departure followed within milliseconds by an
// arrival (measured on Linux, 2026-09-28); the ribbon stays open with nothing pending.
func TestASpuriousDepartureIsAbsorbed(t *testing.T) {
	t.Parallel()
	s := opened(t).Left(at(Rest + time.Millisecond)).Arrived(at(Rest + 6*time.Millisecond))
	if !s.Open() {
		t.Error("a false departure closed the ribbon")
	}
	wantNothingDue(t, s)
}

// FR-616: while held (a drag, a menu, a panel) nothing collapses; released, the second counts from then.
func TestNothingCollapsesWhileHeld(t *testing.T) {
	t.Parallel()
	held := opened(t).Held().Left(at(5 * time.Second))
	wantNothingDue(t, held)
	if !held.At(at(time.Minute)).Open() {
		t.Error("a held ribbon collapsed")
	}
	released := at(8 * time.Second)
	wantDue(t, held.Released(released), released.Add(Away))
	// A departure pending when the hold began is dropped by it.
	wantNothingDue(t, opened(t).Left(at(5*time.Second)).Held())
}

// A panel shows the ribbon's window in full by command: held open while it stands; collapsing Away
// after it closes when the pointer is off it.
func TestAPanelHoldsTheRibbonOpen(t *testing.T) {
	t.Parallel()
	s := State{}.Held().Expanded(start)
	if !s.Open() {
		t.Error("a panel did not open the ribbon")
	}
	wantNothingDue(t, s)
	closed := at(30 * time.Second)
	wantDue(t, s.Released(closed), closed.Add(Away))
	wantNothingDue(t, s.Arrived(at(time.Second)).Released(closed))
}

// FR-613: unpinning leaves the ribbon open, collapsing Away later unless the pointer is on it.
func TestUnpinningCollapsesOnlyOnceThePointerIsAway(t *testing.T) {
	t.Parallel()
	s := Unpinned(start, false)
	if !s.Open() {
		t.Error("unpinning collapsed the ribbon at once")
	}
	wantDue(t, s, at(Away))
	wantNothingDue(t, Unpinned(start, true))
}

func TestOperationsLeaveTheReceiverAsItWas(t *testing.T) {
	t.Parallel()
	s := opened(t)
	_ = s.Left(start)
	_ = s.Held()
	if !s.Open() || !s.inside || s.held {
		t.Errorf("the receiver changed: %+v", s)
	}
	wantNothingDue(t, s)
}
