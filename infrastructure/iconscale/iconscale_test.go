package iconscale

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// twoByTwo is a small image whose averages are easy to state: red above, blue below.
func twoByTwo() *image.NRGBA {
	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	source.Set(0, 0, color.NRGBA{R: 200, A: 255})
	source.Set(1, 0, color.NRGBA{R: 100, A: 255})
	source.Set(0, 1, color.NRGBA{B: 40, A: 255})
	source.Set(1, 1, color.NRGBA{B: 60, A: 255})
	return source
}

func TestTheIconIsAveragedDown(t *testing.T) {
	t.Parallel()
	if got := Down(twoByTwo(), 1).NRGBAAt(0, 0); got != (color.NRGBA{R: 75, B: 25, A: 255}) {
		t.Errorf("got %+v", got)
	}
	if got := Down(twoByTwo(), 2).NRGBAAt(1, 0); got != (color.NRGBA{R: 100, A: 255}) {
		t.Errorf("at the same size a pixel changed: %+v", got)
	}
}

// Growing never leaves a pixel without a source.
func TestAnIconGrownStillHasAColourEverywhere(t *testing.T) {
	t.Parallel()
	grown := Down(twoByTwo(), 4)
	if got := grown.NRGBAAt(3, 3); got != (color.NRGBA{B: 60, A: 255}) {
		t.Errorf("got %+v", got)
	}
}

func TestAPNGIsReadAndAnythingElseRefused(t *testing.T) {
	t.Parallel()
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, twoByTwo()); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(encoded.Bytes()); err != nil {
		t.Error(err)
	}
	if _, err := Decode([]byte("not a png")); err == nil {
		t.Error("something that is not a PNG was read")
	}
}
