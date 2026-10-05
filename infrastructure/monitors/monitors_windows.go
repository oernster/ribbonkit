package monitors

// On Windows the displays are read from Win32 and coordinates are physical pixels. Measured on
// 2026-09-27 from the test binary, which declares no DPI awareness: a display at 250 percent
// reported a work area 3840 pixels wide with a DPI of 240. The application's manifest declares
// per-monitor awareness as well.

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/oernster/ribbonkit/domain/placement"
)

// Win32 constants, named as the headers name them.
const (
	monitorInfoPrimary = 0x1 // MONITORINFOF_PRIMARY
	mdtEffectiveDPI    = 0   // MDT_EFFECTIVE_DPI
	deviceNameLength   = 32  // CCHDEVICENAME
	enumerationGoesOn  = 1   // TRUE, returned from the callback to continue
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	shcore                  = windows.NewLazySystemDLL("shcore.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW     = user32.NewProc("GetMonitorInfoW")
	procGetDpiForMonitor    = shcore.NewProc("GetDpiForMonitor")
)

// rect is Win32 RECT.
type rect struct{ Left, Top, Right, Bottom int32 }

// monitorInfoEx is Win32 MONITORINFOEXW.
type monitorInfoEx struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
	Device  [deviceNameLength]uint16
}

// Monitors is the application's Monitors port.
type Monitors struct{}

// The enumeration callback is made once, since Windows limits how many a process may make; each
// enumeration collects into handles under the lock.
var (
	enumeration sync.Mutex
	handles     []uintptr
	callback    = windows.NewCallback(func(monitor, _, _, _ uintptr) uintptr {
		handles = append(handles, monitor)
		return enumerationGoesOn
	})
)

// Monitors answers every display attached now.
func (Monitors) Monitors() ([]placement.Monitor, error) {
	enumeration.Lock()
	handles = nil
	ok, _, callErr := procEnumDisplayMonitors.Call(0, 0, callback, 0)
	found := handles
	enumeration.Unlock()
	if ok == 0 {
		return nil, fmt.Errorf("EnumDisplayMonitors: %w", callErr)
	}
	out := make([]placement.Monitor, 0, len(found))
	for _, handle := range found {
		monitor, err := describe(handle)
		if err != nil {
			return nil, err
		}
		out = append(out, monitor)
	}
	return out, nil
}

// describe reads one monitor's device name, work area, primary flag and effective DPI.
func describe(handle uintptr) (placement.Monitor, error) {
	info := monitorInfoEx{Size: uint32(unsafe.Sizeof(monitorInfoEx{}))}
	if ok, _, err := procGetMonitorInfoW.Call(handle, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return placement.Monitor{}, fmt.Errorf("GetMonitorInfoW: %w", err)
	}
	var dpiX, dpiY uint32
	if result, _, _ := procGetDpiForMonitor.Call(handle, mdtEffectiveDPI,
		uintptr(unsafe.Pointer(&dpiX)), uintptr(unsafe.Pointer(&dpiY))); result != 0 {
		return placement.Monitor{}, fmt.Errorf("GetDpiForMonitor answered 0x%x", result)
	}
	return placement.Monitor{
		Device:  windows.UTF16ToString(info.Device[:]),
		Work:    placement.Rect{Left: int(info.Work.Left), Top: int(info.Work.Top), Right: int(info.Work.Right), Bottom: int(info.Work.Bottom)},
		DPI:     int(dpiX),
		Primary: info.Flags&monitorInfoPrimary != 0,
	}, nil
}
