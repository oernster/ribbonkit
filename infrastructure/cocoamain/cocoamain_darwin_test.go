package cocoamain

import (
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(Serve(m.Run)) }

func TestDoRunsTheWorkAndWaitsForIt(t *testing.T) {
	ran := false
	Do(func() { ran = true })
	if !ran {
		t.Error("Do returned before the work ran")
	}
}

func TestAPanicOnTheMainThreadReachesTheCaller(t *testing.T) {
	defer func() {
		if failure := recover(); failure == nil || !strings.Contains(failure.(string), "boom") {
			t.Errorf("recovered %v", failure)
		}
	}()
	Do(func() { panic("boom") })
}

func TestNestedWorkOnTheMainThreadRunsAtOnce(t *testing.T) {
	inner := false
	Do(func() { Do(func() { inner = true }) })
	if !inner {
		t.Error("work asked for on the main thread from the main thread did not run")
	}
}
