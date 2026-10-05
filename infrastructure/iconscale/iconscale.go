// Package iconscale makes small copies of the application's icon by averaging. The tray on Linux
// sends one; the Linux build installs a set at the sizes the icon theme asks for. The master is far
// larger than any of them.
package iconscale

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

// Decode reads a PNG.
func Decode(encoded []byte) (image.Image, error) {
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("reading the icon: %w", err)
	}
	return decoded, nil
}

// Down answers source averaged down to size pixels square: each pixel the mean of the source pixels
// it covers, without premultiplied alpha.
func Down(source image.Image, size int) *image.NRGBA {
	bounds := source.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	for row := range size {
		top, bottom := span(bounds.Min.Y, bounds.Dy(), row, size)
		for column := range size {
			left, right := span(bounds.Min.X, bounds.Dx(), column, size)
			out.SetNRGBA(column, row, average(source, left, top, right, bottom))
		}
	}
	return out
}

// span answers the source pixels, from start over length, that target pixel index of count covers:
// never empty, so a target pixel always has a source.
func span(start, length, index, count int) (from, to int) {
	from = start + index*length/count
	to = start + (index+1)*length/count
	if to <= from {
		to = from + 1
	}
	return from, to
}

// average answers the mean colour of the rectangle.
func average(source image.Image, left, top, right, bottom int) color.NRGBA {
	var red, green, blue, alpha, count uint64
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			red += uint64(pixel.R)
			green += uint64(pixel.G)
			blue += uint64(pixel.B)
			alpha += uint64(pixel.A)
			count++
		}
	}
	return color.NRGBA{R: uint8(red / count), G: uint8(green / count), B: uint8(blue / count), A: uint8(alpha / count)}
}
