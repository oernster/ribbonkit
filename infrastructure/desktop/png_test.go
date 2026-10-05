//go:build darwin || linux

package desktop

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

// encodePNG answers picture as a PNG, the form the tray is handed its icon in. The tests make their
// own pictures, so the kit reads no application's icon from disk.
func encodePNG(t *testing.T, picture image.Image) []byte {
	t.Helper()
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, picture); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}
