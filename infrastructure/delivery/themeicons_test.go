package delivery

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// writeMaster writes a small master icon into dir, answering its path.
func writeMaster(t *testing.T, dir string) string {
	t.Helper()
	master := filepath.Join(dir, "master.png")
	file, err := os.Create(master)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, image.NewNRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	return master
}

// Every theme size is written from the master, each at its own size.
func TestEveryThemeSizeIsWritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out := filepath.Join(dir, "icons")
	if err := ThemeIcons([]string{"-in", writeMaster(t, dir), "-out", out, "-name", "ribbon"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, size := range themeSizes {
		written, err := os.Open(filepath.Join(out, fmt.Sprintf("ribbon_%d.png", size)))
		if err != nil {
			t.Fatal(err)
		}
		config, err := png.DecodeConfig(written)
		_ = written.Close()
		if err != nil || config.Width != size || config.Height != size {
			t.Errorf("size %d: %dx%d (%v)", size, config.Width, config.Height, err)
		}
	}
}

func TestThemeIconsRefusesWhatItCannotUse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	master := writeMaster(t, dir)
	notAPicture := filepath.Join(dir, "notes.png")
	if err := os.WriteFile(notAPicture, []byte("not a picture"), fileMode); err != nil {
		t.Fatal(err)
	}
	if err := ThemeIcons([]string{"-in", "x.png"}, io.Discard); !errors.Is(err, errThemeIconFlags) {
		t.Errorf("without -out and -name: %v", err)
	}
	cases := map[string][]string{
		"an unknown flag":         {"-colour", "blue"},
		"a missing master":        {"-in", filepath.Join(dir, "absent.png"), "-out", dir, "-name", "r"},
		"a master not a PNG":      {"-in", notAPicture, "-out", dir, "-name", "r"},
		"a folder that is a file": {"-in", master, "-out", master, "-name", "r"},
		"a size it cannot write":  {"-in", master, "-out", dir, "-name", filepath.Join("absent", "r")},
	}
	for name, args := range cases {
		if err := ThemeIcons(args, io.Discard); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
