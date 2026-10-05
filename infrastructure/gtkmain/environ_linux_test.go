package gtkmain

import (
	"os"
	"testing"
)

// With nothing chosen, the DMABUF renderer is turned off.
func TestTheDMABUFRendererIsTurnedOff(t *testing.T) {
	t.Setenv(dmabufVariable, "")
	_ = os.Unsetenv(dmabufVariable)
	AvoidDMABUF()
	if got := os.Getenv(dmabufVariable); got != dmabufOff {
		t.Errorf("%s is %q", dmabufVariable, got)
	}
}

// A value the environment already holds is kept, so the renderer can be turned back on.
func TestAChosenDMABUFSettingIsKept(t *testing.T) {
	const chosen = "0"
	t.Setenv(dmabufVariable, chosen)
	AvoidDMABUF()
	if got := os.Getenv(dmabufVariable); got != chosen {
		t.Errorf("%s is %q, not the %q chosen", dmabufVariable, got, chosen)
	}
}
