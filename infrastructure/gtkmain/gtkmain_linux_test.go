package gtkmain

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(ServeTests(m.Run)) }

// Do runs its work on the loop and returns only once it has.
func TestDoRunsTheWorkAndWaitsForIt(t *testing.T) {
	ran := false
	Do(func() { ran = true })
	if !ran {
		t.Error("Do returned before its work ran")
	}
}

// A panic on the loop reaches the caller rather than ending the process.
func TestAPanicOnTheLoopReachesTheCaller(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("the panic did not reach the caller")
		}
	}()
	Do(func() { panic("planted") })
}

// The tests themselves run on X11, as the application does.
func TestTheBackendIsX11(t *testing.T) {
	if got := os.Getenv(backendVariable); got != backendX11 {
		t.Errorf("GDK_BACKEND is %q", got)
	}
}
