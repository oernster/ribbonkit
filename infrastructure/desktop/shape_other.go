//go:build !windows

package desktop

import "github.com/oernster/ribbonkit/domain/placement"

// Shape leaves the window a rectangle: on macOS and Linux it is not cut to the ribbon and its pull
// out (FR-913).
func Shape(Window, []placement.Rect) error { return nil }
