package desktop

import (
	"io"

	"github.com/oernster/ribbonkit/application/shell"
	"github.com/oernster/ribbonkit/domain/placement"
)

// Desktop is the shell port on every platform: the methods each platform's half declares, plus the
// operations below, which are the package's own functions reached through the port.
var _ shell.Desktop = (*Desktop)(nil)

// FindRibbon is the package's FindRibbon.
func (*Desktop) FindRibbon(class, name string) (Window, error) { return FindRibbon(class, name) }

// HideFromTaskbar is the package's HideFromTaskbar.
func (*Desktop) HideFromTaskbar(ribbon Window) error { return HideFromTaskbar(ribbon) }

// KeepOnDisplays is the package's KeepOnDisplays.
func (*Desktop) KeepOnDisplays(ribbon Window, log io.Writer) error {
	return KeepOnDisplays(ribbon, log)
}

// Place is the package's Place.
func (*Desktop) Place(ribbon Window, at placement.Point, size placement.Size) error {
	return Place(ribbon, at, size)
}

// Position is the package's Position.
func (*Desktop) Position(ribbon Window) (placement.Point, error) { return Position(ribbon) }

// Shape is the package's Shape.
func (*Desktop) Shape(ribbon Window, parts []placement.Rect) error { return Shape(ribbon, parts) }

// SetTabFrame is the package's SetTabFrame.
func (*Desktop) SetTabFrame(ribbon Window, tab bool) error { return SetTabFrame(ribbon, tab) }

// Cursor is the package's Cursor.
func (*Desktop) Cursor() (placement.Point, bool) { return Cursor() }

// DragThreshold is the package's DragThreshold.
func (*Desktop) DragThreshold() placement.Size { return DragThreshold() }

// ToolkitScale is the package's ToolkitScale.
func (*Desktop) ToolkitScale() int { return ToolkitScale() }

// PixelsPerDIP is the package's PixelsPerDIP.
func (*Desktop) PixelsPerDIP(pageRatio float64, toolkitScale int) float64 {
	return PixelsPerDIP(pageRatio, toolkitScale)
}

// OpenInBrowser is the package's OpenInBrowser.
func (*Desktop) OpenInBrowser(address string) error { return OpenInBrowser(address) }
