package desktop

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/oernster/ribbonkit/domain/placement"
)

// FindRibbon answers the window of class, the class name the ribbon's window is created with. The
// name is for the desktops that find it by its title instead.
func FindRibbon(class, _ string) (Window, error) {
	for range findAttempts {
		if handle, _, _ := procFindWindow.Call(utf16Pointer(class), 0); handle != 0 {
			return Window(handle), nil
		}
		time.Sleep(findPause)
	}
	return 0, fmt.Errorf("%w: class %s", ErrRibbonNotFound, class)
}

// HideFromTaskbar makes the ribbon a tool window, which has no taskbar button (FR-101). Wails creates
// its window with WS_EX_APPWINDOW, which forces a button, so that flag is taken off. It is done
// while the window is hidden, since the taskbar reads the style when a window is shown.
func HideFromTaskbar(ribbon Window) error {
	style, _, _ := procGetWindowLongPtr.Call(uintptr(ribbon), uintptr(exStyleIndex))
	// SetWindowLongPtr answers the previous style, never zero here since Wails sets WS_EX_APPWINDOW;
	// zero is the failure.
	if previous, _, err := procSetWindowLongPtr.Call(uintptr(ribbon), uintptr(exStyleIndex), style&^wsExAppWindow|wsExToolWindow); previous == 0 {
		return fmt.Errorf("changing the ribbon's window style: %w", err)
	}
	return nil
}

// exStyleIndex is GWL_EXSTYLE held in a variable, since a negative constant cannot convert to the
// unsigned argument the call takes.
var exStyleIndex int32 = gwlExStyle

// Place moves and sizes the ribbon in physical pixels without raising or activating it (FR-405).
func Place(ribbon Window, at placement.Point, size placement.Size) error {
	ok, _, err := procSetWindowPos.Call(uintptr(ribbon), 0,
		uintptr(at.X), uintptr(at.Y), uintptr(size.Width), uintptr(size.Height),
		swpNoZOrder|swpNoActivate|swpFrameChanged)
	if ok == 0 {
		return fmt.Errorf("placing the ribbon: %w", err)
	}
	return nil
}

// Position answers the ribbon's top-left corner in physical pixels.
func Position(ribbon Window) (placement.Point, error) {
	var bounds rect
	if ok, _, err := procGetWindowRect.Call(uintptr(ribbon), uintptr(unsafe.Pointer(&bounds))); ok == 0 {
		return placement.Point{}, fmt.Errorf("reading the ribbon's position: %w", err)
	}
	return placement.Point{X: int(bounds.left), Y: int(bounds.top)}, nil
}

// noWindowRegion is what GetWindowRgn answers for a window with no shape set (ERROR): all of its
// rectangle is the window.
const noWindowRegion = 0

// pointerInside answers whether the pointer is on the ribbon's window now (FR-615, FR-616): inside
// its rectangle and inside the shape it is cut to, so a pointer resting where the shape cuts the
// window away (beside a vertical ribbon's pull out, measured 2026-09-29) is off it (FR-913).
func pointerInside(ribbon Window) (bool, error) {
	var cursor point
	if ok, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor))); ok == 0 {
		return false, fmt.Errorf("reading the pointer: %w", err)
	}
	var bounds rect
	if ok, _, err := procGetWindowRect.Call(uintptr(ribbon), uintptr(unsafe.Pointer(&bounds))); ok == 0 {
		return false, fmt.Errorf("reading the ribbon's rectangle: %w", err)
	}
	if cursor.x < bounds.left || cursor.x >= bounds.right || cursor.y < bounds.top || cursor.y >= bounds.bottom {
		return false, nil
	}
	shape, _, _ := procCreateRectRgn.Call(0, 0, 0, 0)
	if shape == 0 {
		return false, fmt.Errorf("reading the ribbon's shape: %w", errNoRegion)
	}
	defer procDeleteObject.Call(shape)
	if kind, _, _ := procGetWindowRgn.Call(uintptr(ribbon), shape); kind == noWindowRegion {
		return true, nil
	}
	// The shape is held in the window's own pixels, from its top-left corner.
	in, _, _ := procPtInRegion.Call(shape, uintptr(cursor.x-bounds.left), uintptr(cursor.y-bounds.top))
	return in != 0, nil
}

// tabStyles are the styles Wails gives its frameless window (style 0x4ca0000, measured 2026-09-28)
// that hold it at least 42 pixels wide; without them it takes the tab's 8 across (FR-614).
const tabStyles = wsCaption | wsSysMenu | wsMinimizeBox

// An overlapped window, as Wails' is, is still held at least 39 pixels tall (SM_CYMINTRACK) without
// those styles, which kept a horizontal ribbon's tab at 39; as a popup it takes 8 (measured
// 2026-09-28). So the tab is a popup as well.

// styleIndex is GWL_STYLE held in a variable, as exStyleIndex is.
var styleIndex int32 = gwlStyle

// SetTabFrame takes tabStyles off the ribbon's window while it wears the tab's frame, which it does
// for as long as it is unpinned (full or tab). It gives them back for a pinned ribbon or a panel,
// which keep the window Wails made. The next Place applies the change.
func SetTabFrame(ribbon Window, tab bool) error {
	// Every window has a style here (WS_CLIPSIBLINGS at least), so zero is a failed read.
	style, _, err := procGetWindowLongPtr.Call(uintptr(ribbon), uintptr(styleIndex))
	if style == 0 {
		return fmt.Errorf("reading the ribbon's frame: %w", err)
	}
	next := style&^wsPopup | tabStyles
	if tab {
		next = style&^tabStyles | wsPopup
	}
	if next == style {
		return nil
	}
	// The previous style is never zero, as above; zero is failure.
	if previous, _, err := procSetWindowLongPtr.Call(uintptr(ribbon), uintptr(styleIndex), next); previous == 0 {
		return fmt.Errorf("changing the ribbon's frame: %w", err)
	}
	return nil
}

// DragThreshold answers how far the pointer must move, in DIP, before a press becomes a drag: the
// Windows drag rectangle at 100 percent scaling (FR-401).
func DragThreshold() placement.Size {
	width, _, _ := procGetSystemMetricsForDpi.Call(smCxDrag, placement.BaseDPI)
	height, _, _ := procGetSystemMetricsForDpi.Call(smCyDrag, placement.BaseDPI)
	return placement.Size{Width: int(width), Height: int(height)}
}
