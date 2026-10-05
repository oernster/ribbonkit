package desktop

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/application/menus"
)

// The icon appears once AppKit's loop has served the request, so its arrival is polled for.
const (
	iconLimit = 2 * time.Second
	iconPause = 20 * time.Millisecond
)

// testIconSize is the side of the small square picture the menu bar is handed.
const testIconSize = 22

// opaqueSquare answers a square picture with every pixel opaque, as an application's icon is.
func opaqueSquare() image.Image {
	picture := image.NewNRGBA(image.Rect(0, 0, testIconSize, testIconSize))
	draw.Draw(picture, picture.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
	return picture
}

// FR-501: the icon stands in the menu bar with its image until Stop takes it out again.
func TestTheIconStandsInTheMenuBarUntilStop(t *testing.T) {
	d := New(testApp, func() []menus.Item { return nil }, io.Discard)
	d.UseIcon(encodePNG(t, opaqueSquare()))
	if err := d.Start(); err != nil {
		t.Fatal(err)
	}
	shown := false
	for deadline := time.Now().Add(iconLimit); time.Now().Before(deadline) && !shown; time.Sleep(iconPause) {
		shown = shownInMenuBar()
	}
	if !shown {
		t.Error("the icon did not appear in the menu bar")
	}
	d.Stop()
	if shownInMenuBar() {
		t.Error("the icon is still in the menu bar after Stop")
	}
}

// A desktop given no image says so rather than showing an empty icon.
func TestAnIconWithNoImageIsRefused(t *testing.T) {
	d := New(testApp, func() []menus.Item { return nil }, io.Discard)
	defer d.Stop()
	if err := d.Start(); !errors.Is(err, errNoIcon) {
		t.Errorf("got %v", err)
	}
}
