package monitors

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#include <stdlib.h>
#include <string.h>

// screen_info is one display as AppKit describes it, in points, with Cocoa's origin at the bottom
// left of the menu-bar display.
typedef struct {
    double x, y, width, height;
    double visibleX, visibleY, visibleWidth, visibleHeight;
    double scale;
    char *name;
} screen_info;

static int screen_count(void) { return (int)[[NSScreen screens] count]; }

// read_screens fills out with up to max displays, the menu-bar display first, answering how many
// it filled. Each name is the caller's to free.
static int read_screens(screen_info *out, int max)
{
    NSArray<NSScreen *> *screens = [NSScreen screens];
    int filled = 0;
    for (NSScreen *screen in screens) {
        if (filled >= max) {
            break;
        }
        NSRect frame = [screen frame], visible = [screen visibleFrame];
        NSString *name = [screen localizedName];
        out[filled] = (screen_info){
            frame.origin.x, frame.origin.y, frame.size.width, frame.size.height,
            visible.origin.x, visible.origin.y, visible.size.width, visible.size.height,
            [screen backingScaleFactor], strdup(name == nil ? "" : [name UTF8String]),
        };
        filled++;
    }
    return filled;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/cocoamain"
)

// On macOS the displays are read from AppKit, in points: device pixels divided by the display's
// backing scale. The ribbon is placed and sized in those same points and AppKit does the scaling,
// so every display is reported at BaseDPI, where one point is one DIP.
//
// AppKit counts upward from the bottom-left corner of the menu-bar display, which it lists first;
// The kit counts downward from a top-left corner, as Windows and GTK do. So a rectangle is
// turned over against the menu-bar display's height: a top edge is that height less the Cocoa top.

// errNoDisplay is answered when AppKit holds no display to read.
var errNoDisplay = errors.New("AppKit holds no display")

// Monitors is the application's Monitors port.
type Monitors struct{}

// display is one display as AppKit describes it: what the application is told, plus the frame and
// scale a person checking the units wants to see.
type display struct {
	monitor  placement.Monitor
	geometry placement.Rect
	scale    float64
}

// cocoaRect is a rectangle as AppKit gives it: a bottom-left corner with a width and height.
type cocoaRect struct{ x, y, width, height float64 }

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

// readDisplays reads every display on AppKit's main thread.
func readDisplays() ([]display, error) {
	var out []display
	var err error
	cocoamain.Do(func() { out, err = describeAll() })
	return out, err
}

// describeAll reads each of AppKit's displays. It runs on the main thread.
func describeAll() ([]display, error) {
	count := int(C.screen_count())
	if count == 0 {
		return nil, errNoDisplay
	}
	buffer := (*C.screen_info)(C.calloc(C.size_t(count), C.size_t(unsafe.Sizeof(C.screen_info{}))))
	defer C.free(unsafe.Pointer(buffer))
	infos := unsafe.Slice(buffer, count)[:int(C.read_screens(buffer, C.int(count)))]
	if len(infos) == 0 {
		return nil, errNoDisplay
	}
	primaryHeight := float64(infos[0].height)
	out := make([]display, 0, len(infos))
	for index, info := range infos {
		name := C.GoString(info.name)
		C.free(unsafe.Pointer(info.name))
		frame := cocoaRect{float64(info.x), float64(info.y), float64(info.width), float64(info.height)}
		visible := cocoaRect{float64(info.visibleX), float64(info.visibleY), float64(info.visibleWidth), float64(info.visibleHeight)}
		out = append(out, display{
			monitor: placement.Monitor{
				Device:  deviceName(index, name),
				Work:    turnedOver(visible, primaryHeight),
				DPI:     placement.BaseDPI,
				Primary: index == 0,
			},
			geometry: turnedOver(frame, primaryHeight),
			scale:    float64(info.scale),
		})
	}
	return out, nil
}

// turnedOver answers r counted downward from the top of a menu-bar display primaryHeight tall.
func turnedOver(r cocoaRect, primaryHeight float64) placement.Rect {
	return placement.Rect{
		Left:   int(r.x),
		Top:    int(primaryHeight - (r.y + r.height)),
		Right:  int(r.x + r.width),
		Bottom: int(primaryHeight - r.y),
	}
}

// deviceName names a display by its place in AppKit's list and the name macOS shows for it.
func deviceName(index int, name string) string {
	if name == "" {
		return fmt.Sprintf("display %d", index+1)
	}
	return fmt.Sprintf("display %d %s", index+1, name)
}
