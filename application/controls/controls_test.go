package controls

import (
	"context"
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/release"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// errPlanted is the failure the fakes answer when told to.
var errPlanted = errors.New("planted")

// host holds the choices as the application would, counting saves; saveErr fails each save after the
// change is kept, as an application's save path does (FR-707).
type host struct {
	choices ribbon.Choices
	saves   int
	saveErr error
}

func (h *host) Ribbon() (ribbon.Choices, arranger.Content) { return h.choices, arranger.Content{} }

func (h *host) ChangeRibbon(edit func(ribbon.Choices) ribbon.Choices) error {
	h.choices = edit(h.choices)
	h.saves++
	return h.saveErr
}

// startup records the sign-in entry.
type startup struct {
	on  bool
	err error
}

func (s *startup) Enabled() (bool, error) { return s.on, s.err }
func (s *startup) Enable() error          { s.on = true; return s.err }
func (s *startup) Disable() error         { s.on = false; return s.err }

// source answers one release.
type source struct{ version string }

func (s source) LatestRelease(context.Context) (release.Info, error) {
	return release.Info{Version: s.version}, nil
}

// controlsOver answers controls over fresh fakes, the running build 1.0.0 with 1.1.0 published.
func controlsOver() (*Controls, *host, *startup) {
	h := &host{choices: ribbon.Defaults()}
	s := &startup{}
	return New(h, s, source{version: "1.1.0"}, release.Build{Version: "1.0.0"}), h, s
}

// Each choice is saved through the host, the others normalised with it.
func TestEachChoiceIsSavedThroughTheHost(t *testing.T) {
	t.Parallel()
	c, h, _ := controlsOver()
	h.choices.Theme = "chartreuse"
	steps := []error{
		c.SetColour(ribbon.Sunset), c.SetOrientation(ribbon.Horizontal), c.SetTheme(ribbon.Dark),
		c.SetAlwaysOnTop(true), c.SetPinned(false), c.SetOpacity(55), c.SkipUpdate("1.1.0"),
	}
	if err := errors.Join(steps...); err != nil {
		t.Fatal(err)
	}
	want := ribbon.Defaults()
	want.Colour, want.Orientation, want.Theme, want.AlwaysOnTop = ribbon.Sunset, ribbon.Horizontal, ribbon.Dark, true
	want.Pinned, want.Opacity, want.SkippedUpdate = false, 55, "1.1.0"
	if h.choices != want || h.saves != len(steps) {
		t.Errorf("choices %+v after %d saves", h.choices, h.saves)
	}
}

// A value a choice does not offer is refused and nothing is saved (FR-602).
func TestAValueNotOfferedIsRefused(t *testing.T) {
	t.Parallel()
	c, h, _ := controlsOver()
	for name, err := range map[string]error{
		"colour":      c.SetColour("plaid"),
		"orientation": c.SetOrientation("diagonal"),
		"theme":       c.SetTheme("sepia"),
		"opacity":     c.SetOpacity(ribbon.MinOpacity - 1),
	} {
		if !errors.Is(err, ribbon.ErrUnknownChoice) {
			t.Errorf("%s answered %v", name, err)
		}
	}
	if h.saves != 0 || h.choices != ribbon.Defaults() {
		t.Errorf("saved %d times, choices %+v", h.saves, h.choices)
	}
}

// A save that fails keeps the change and is answered (FR-707).
func TestAFailedSaveIsAnswered(t *testing.T) {
	t.Parallel()
	c, h, _ := controlsOver()
	h.saveErr = errPlanted
	if err := c.SetPinned(false); !errors.Is(err, errPlanted) || h.choices.Pinned {
		t.Errorf("answered %v with pinned %v", err, h.choices.Pinned)
	}
}

// The sign-in entry is read, written and removed through the port, its failures answered (FR-605).
func TestStartAtSignInGoesThroughTheEntry(t *testing.T) {
	t.Parallel()
	c, _, s := controlsOver()
	if err := c.SetStartWithWindows(true); err != nil || !s.on {
		t.Fatalf("enable: %v, on %v", err, s.on)
	}
	if on, err := c.StartWithWindows(); err != nil || !on {
		t.Errorf("read %v, %v", on, err)
	}
	if err := c.SetStartWithWindows(false); err != nil || s.on {
		t.Errorf("disable: %v, on %v", err, s.on)
	}
	s.err = errPlanted
	if err := c.SetStartWithWindows(true); !errors.Is(err, errPlanted) {
		t.Errorf("answered %v", err)
	}
}

// The automatic check is never offered the release the user skipped; a manual one ignores the skip
// (FR-509).
func TestTheSkippedReleaseIsOfferedOnlyWhenAskedFor(t *testing.T) {
	t.Parallel()
	c, _, _ := controlsOver()
	if status := c.CheckForUpdate(context.Background(), false); !status.Available || status.Latest != "1.1.0" {
		t.Fatalf("before skipping: %+v", status)
	}
	if err := c.SkipUpdate("1.1.0"); err != nil {
		t.Fatal(err)
	}
	if status := c.CheckForUpdate(context.Background(), false); status.Available {
		t.Errorf("the skipped release was offered: %+v", status)
	}
	if status := c.CheckForUpdate(context.Background(), true); !status.Available {
		t.Errorf("a manual check did not offer it: %+v", status)
	}
}
