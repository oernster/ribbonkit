package monitors

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
*/
import "C"

import (
	"errors"
	"fmt"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/gtkmain"
)

// On Linux the displays are read from GDK, whose coordinates are GTK's own units: device pixels
// divided by the display's scale factor. The ribbon is placed and sized in those same units and
// GTK does the scaling, so every display is reported at BaseDPI, where one unit is one DIP.

// errNoDisplay is answered when GDK holds no display to read.
var errNoDisplay = errors.New("GDK holds no display")

// Monitors is the application's Monitors port.
type Monitors struct{}

// display is one display as GDK describes it: what the application is told, plus the geometry and
// scale factor a person checking the units wants to see.
type display struct {
	monitor  placement.Monitor
	geometry placement.Rect
	scale    int
}

// Monitors answers every display attached now.
func (Monitors) Monitors() ([]placement.Monitor, error) {
	displays, err := readDisplays()
	if err != nil {
		return nil, err
	}
	out := make([]placement.Monitor, 0, len(displays))
	for _, each := range displays {
		out = append(out, each.monitor)
	}
	return out, nil
}

// readDisplays reads every display on GTK's main loop.
func readDisplays() ([]display, error) {
	var out []display
	var err error
	gtkmain.Do(func() { out, err = describeAll() })
	return out, err
}

// describeAll reads each of GDK's displays. It runs on GTK's main loop.
func describeAll() ([]display, error) {
	gdk := C.gdk_display_get_default()
	if gdk == nil {
		return nil, errNoDisplay
	}
	count := int(C.gdk_display_get_n_monitors(gdk))
	out := make([]display, 0, count)
	for index := range count {
		monitor := C.gdk_display_get_monitor(gdk, C.int(index))
		if monitor == nil {
			continue
		}
		var work, geometry C.GdkRectangle
		C.gdk_monitor_get_workarea(monitor, &work)
		C.gdk_monitor_get_geometry(monitor, &geometry)
		model := ""
		if name := C.gdk_monitor_get_model(monitor); name != nil {
			model = C.GoString(name)
		}
		out = append(out, display{
			monitor: placement.Monitor{
				Device:  deviceName(index, model),
				Work:    rectOf(work),
				DPI:     placement.BaseDPI,
				Primary: C.gdk_monitor_is_primary(monitor) != 0,
			},
			geometry: rectOf(geometry),
			scale:    int(C.gdk_monitor_get_scale_factor(monitor)),
		})
	}
	return out, nil
}

// deviceName names a display by its place in GDK's list and its model, which GDK may not know. GTK
// 3 offers no connector name, so a display is known by where it stands in the list.
func deviceName(index int, model string) string {
	if model == "" {
		return fmt.Sprintf("display %d", index+1)
	}
	return fmt.Sprintf("display %d %s", index+1, model)
}

// rectOf turns a GDK rectangle, a corner with a width and height, into a placement.Rect.
func rectOf(r C.GdkRectangle) placement.Rect {
	return placement.Rect{Left: int(r.x), Top: int(r.y), Right: int(r.x + r.width), Bottom: int(r.y + r.height)}
}
