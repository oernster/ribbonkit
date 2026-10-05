// Package controls is every ribbon's own use cases over the ribbon's choices: choosing its colour,
// orientation, theme, opacity, Always on top and the pin; starting at sign-in; the update check and
// the release the user skipped. An application embeds Controls in its service, so these are its
// service's own. It hands Controls the arranger's host, so every change is saved through the
// application's one save path. FR numbers are TimeRibbon's.
package controls

import (
	"context"
	"fmt"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/release"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// Startup is the entry that starts the application at sign-in (FR-605). The system holds the
// answer, not the settings file, so the setting and setup cannot disagree (FR-805).
type Startup interface {
	Enabled() (bool, error)
	Enable() error
	Disable() error
}

// Controls runs the ribbon's own use cases.
type Controls struct {
	host     arranger.Host
	startup  Startup
	releases release.Source
	build    release.Build
}

// New answers the controls over host, the sign-in entry, the release source and the running build.
func New(host arranger.Host, startup Startup, releases release.Source, build release.Build) *Controls {
	return &Controls{host: host, startup: startup, releases: releases, build: build}
}

// SetColour chooses the colour scheme every cell is drawn in (FR-611).
func (c *Controls) SetColour(colour ribbon.Colour) error {
	return choose(c, colour, func(choices *ribbon.Choices) *ribbon.Colour { return &choices.Colour })
}

// SetOrientation chooses horizontal or vertical (FR-103).
func (c *Controls) SetOrientation(orientation ribbon.Orientation) error {
	return choose(c, orientation, func(choices *ribbon.Choices) *ribbon.Orientation { return &choices.Orientation })
}

// SetTheme chooses system, light or dark (FR-606).
func (c *Controls) SetTheme(theme ribbon.Theme) error {
	return choose(c, theme, func(choices *ribbon.Choices) *ribbon.Theme { return &choices.Theme })
}

// SetAlwaysOnTop turns Always on top on or off (FR-505).
func (c *Controls) SetAlwaysOnTop(on bool) error {
	return choose(c, on, func(choices *ribbon.Choices) *bool { return &choices.AlwaysOnTop })
}

// SetPinned pins or unpins the ribbon (FR-613).
func (c *Controls) SetPinned(on bool) error {
	return choose(c, on, func(choices *ribbon.Choices) *bool { return &choices.Pinned })
}

// SetOpacity chooses how opaque the window is drawn, in percent (FR-622). A value outside
// ribbon.MinOpacity to ribbon.MaxOpacity is refused and changes nothing.
func (c *Controls) SetOpacity(percent int) error {
	return choose(c, percent, func(choices *ribbon.Choices) *int { return &choices.Opacity })
}

// SkipUpdate keeps version as the release the automatic check never offers again (FR-509).
func (c *Controls) SkipUpdate(version string) error {
	return choose(c, version, func(choices *ribbon.Choices) *string { return &choices.SkippedUpdate })
}

// CheckForUpdate answers the update status of the running build (FR-509). The automatic check passes
// manual false and is never offered the release the user skipped; a manual check is asked for, so it
// ignores the skip.
func (c *Controls) CheckForUpdate(ctx context.Context, manual bool) release.Status {
	choices, _ := c.host.Ribbon()
	return release.Check(ctx, c.releases, c.build, choices.SkippedUpdate, manual)
}

// StartWithWindows answers whether the sign-in entry is present (FR-605).
func (c *Controls) StartWithWindows() (bool, error) { return c.startup.Enabled() }

// SetStartWithWindows writes or removes the sign-in entry (FR-605).
func (c *Controls) SetStartWithWindows(on bool) error {
	if on {
		return c.startup.Enable()
	}
	return c.startup.Disable()
}

// choose sets the choice field picks to value through the host's save path, the other choices
// normalised with it. A value the choice does not offer, which normalising would replace, is refused
// and nothing changes (FR-602).
func choose[T comparable](c *Controls, value T, field func(*ribbon.Choices) *T) error {
	current, _ := c.host.Ribbon()
	if !offered(current, value, field) {
		return fmt.Errorf("%w: %v", ribbon.ErrUnknownChoice, value)
	}
	return c.host.ChangeRibbon(func(choices ribbon.Choices) ribbon.Choices {
		next := choices.Normalised()
		*field(&next) = value
		return next
	})
}

// offered answers whether value survives normalising when chosen over choices.
func offered[T comparable](choices ribbon.Choices, value T, field func(*ribbon.Choices) *T) bool {
	next := choices.Normalised()
	*field(&next) = value
	normalised := next.Normalised()
	return *field(&normalised) == value
}
