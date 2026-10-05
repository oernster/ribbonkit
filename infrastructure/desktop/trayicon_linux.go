package desktop

import (
	"image/color"

	"github.com/oernster/ribbonkit/infrastructure/iconscale"
)

// trayIconSize is the edge in pixels of the icon handed to the tray host. The host scales it to the
// panel; the master is too large to send, at four bytes a pixel.
const trayIconSize = 64

// argbBytes is the size of one pixel as the specification sends it: alpha, red, green and blue.
const argbBytes = 4

// pixmap is one icon image as the StatusNotifierItem specification carries it: a width, a height
// and ARGB32 pixels in network byte order, row by row.
type pixmap struct {
	Width  int32
	Height int32
	Data   []byte
}

// pixmapOf decodes a PNG and averages it down to size pixels square.
func pixmapOf(encoded []byte, size int) (pixmap, error) {
	source, err := iconscale.Decode(encoded)
	if err != nil {
		return pixmap{}, err
	}
	small := iconscale.Down(source, size)
	data := make([]byte, 0, size*size*argbBytes)
	for row := range size {
		for column := range size {
			data = append(data, argb(small.NRGBAAt(column, row))...)
		}
	}
	return pixmap{Width: int32(size), Height: int32(size), Data: data}, nil
}

// argb answers a pixel's bytes in the specification's order.
func argb(pixel color.NRGBA) []byte { return []byte{pixel.A, pixel.R, pixel.G, pixel.B} }
