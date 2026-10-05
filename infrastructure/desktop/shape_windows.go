package desktop

import (
	"errors"
	"fmt"

	"github.com/oernster/ribbonkit/domain/placement"
)

// errNoRegion is a region Windows would not make.
var errNoRegion = errors.New("no region was made")

// Shape limits the ribbon's window to parts, in the window's own pixels (FR-913): wherever the
// window's rectangle lies outside them the desktop shows and the pointer passes through. Windows owns
// the region once it is set, so it is freed here only when setting it fails.
func Shape(ribbon Window, parts []placement.Rect) error {
	shape, err := regionOf(parts)
	if err != nil {
		return fmt.Errorf("shaping the ribbon: %w", err)
	}
	if ok, _, err := procSetWindowRgn.Call(uintptr(ribbon), shape, 1); ok == 0 {
		procDeleteObject.Call(shape)
		return fmt.Errorf("shaping the ribbon: %w", err)
	}
	return nil
}

// regionOf answers one region covering every part; none, the whole window, where there are no parts.
func regionOf(parts []placement.Rect) (uintptr, error) {
	var shape uintptr
	for _, part := range parts {
		next, _, _ := procCreateRectRgn.Call(uintptr(part.Left), uintptr(part.Top), uintptr(part.Right), uintptr(part.Bottom))
		if next == 0 {
			if shape != 0 {
				procDeleteObject.Call(shape)
			}
			return 0, fmt.Errorf("%w: %+v", errNoRegion, part)
		}
		if shape == 0 {
			shape = next
			continue
		}
		combined, _, _ := procCombineRgn.Call(shape, shape, next, rgnOr)
		procDeleteObject.Call(next)
		if combined == 0 {
			procDeleteObject.Call(shape)
			return 0, fmt.Errorf("%w: joining %+v", errNoRegion, part)
		}
	}
	return shape, nil
}
