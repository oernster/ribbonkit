// Package windowtest holds what an application's facade tests stand in for the window with: a Control
// recording, in order, what the application's own half asked of the window. It is test support, never
// imported by anything that ships.
package windowtest

import (
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/ui/window"
)

// Control stands in for the window's Control, recording each call made of it.
type Control struct {
	// Calls are the calls that fit, redraw or count the page measured, in order: Refitted,
	// ContentChanged, Redraw, Redrawn and PageMeasured, each by its name.
	Calls []string
	// Reported is each failure reported, as its doing; Panels each panel shown; Colours and Turned
	// each colour and orientation chosen.
	Reported, Panels, Colours, Turned []string
	// Showing is what Shown answers; ChoiceErr what SetColour and SetOrientation answer.
	Showing   window.Shown
	ChoiceErr error
}

func (c *Control) record(call string) { c.Calls = append(c.Calls, call) }

// Count answers how often call was made.
func (c *Control) Count(call string) int {
	count := 0
	for _, each := range c.Calls {
		if each == call {
			count++
		}
	}
	return count
}

// Refitted records the call and answers err, as the window answers the change it fitted after.
func (c *Control) Refitted(err error) error {
	c.record("Refitted")
	return err
}

// ContentChanged records the call.
func (c *Control) ContentChanged() { c.record("ContentChanged") }

// Redraw records the call.
func (c *Control) Redraw() { c.record("Redraw") }

// Redrawn records the call and answers err, as the window answers the change it redrew after.
func (c *Control) Redrawn(err error) error {
	c.record("Redrawn")
	return err
}

// PageMeasured records the call.
func (c *Control) PageMeasured() { c.record("PageMeasured") }

// Report records doing where err is a failure.
func (c *Control) Report(doing string, err error) {
	if err != nil {
		c.Reported = append(c.Reported, doing)
	}
}

// ShowPanel records the panel shown.
func (c *Control) ShowPanel(panel string) { c.Panels = append(c.Panels, panel) }

// Shown answers Showing.
func (c *Control) Shown() window.Shown { return c.Showing }

// Offered greys every item, so a test reads which menu passed through it.
func (c *Control) Offered(items []menus.Item) []menus.Item {
	greyed := make([]menus.Item, len(items))
	for index, item := range items {
		item.Disabled = true
		greyed[index] = item
	}
	return greyed
}

// SetColour records the colour chosen and answers ChoiceErr.
func (c *Control) SetColour(colour string) error {
	c.Colours = append(c.Colours, colour)
	return c.ChoiceErr
}

// SetOrientation records the orientation chosen and answers ChoiceErr.
func (c *Control) SetOrientation(orientation string) error {
	c.Turned = append(c.Turned, orientation)
	return c.ChoiceErr
}
