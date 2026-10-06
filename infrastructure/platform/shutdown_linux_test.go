package platform

import (
	"testing"

	"github.com/godbus/dbus/v5"
)

// Only logind announcing that a shutdown or restart is starting ends the ribbon. The same signal
// with false (a cancelled shutdown) does not; nor does anything else on the bus.
func TestOnlyAShutdownStartingEndsTheRibbon(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		signal *dbus.Signal
		want   bool
	}{
		{"starting", &dbus.Signal{Name: prepareForShutdown, Body: []any{true}}, true},
		{"cancelled", &dbus.Signal{Name: prepareForShutdown, Body: []any{false}}, false},
		{"another signal", &dbus.Signal{Name: login1Manager + ".PrepareForSleep", Body: []any{true}}, false},
		{"no body", &dbus.Signal{Name: prepareForShutdown}, false},
		{"not a truth value", &dbus.Signal{Name: prepareForShutdown, Body: []any{"true"}}, false},
		{"nothing", nil, false},
	}
	for _, each := range cases {
		if got := shutdownStarting(each.signal); got != each.want {
			t.Errorf("%s: shutdownStarting %v, want %v", each.name, got, each.want)
		}
	}
}
