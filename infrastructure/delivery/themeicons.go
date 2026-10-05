package delivery

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/oernster/ribbonkit/infrastructure/iconscale"
)

// themeSizes are the hicolor sizes installed: the freedesktop icon theme's usual set.
var themeSizes = []int{16, 24, 32, 48, 64, 128, 256, 512}

// errThemeIconFlags is answered when ThemeIcons is not told its master, its folder and its name.
var errThemeIconFlags = errors.New("-in, -out and -name are all needed")

// ThemeIcons writes the application's icon at each size the Linux icon theme asks for, as
// <name>_<size>.png, for the Flatpak build to install under hicolor. It reads the one master icon,
// so no size is kept as a second copy. Its flags:
//
//	-in build/appicon.png -out build/linux/icons -name timeribbon
func ThemeIcons(args []string, errOut io.Writer) error {
	flags := flag.NewFlagSet("linuxicons", flag.ContinueOnError)
	flags.SetOutput(errOut)
	in := flags.String("in", "", "the master icon, a PNG")
	out := flags.String("out", "", "the folder to write the sizes into")
	name := flags.String("name", "", "the file name each size starts with")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *in == "" || *out == "" || *name == "" {
		return errThemeIconFlags
	}
	encoded, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	master, err := iconscale.Decode(encoded)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, folderMode); err != nil {
		return err
	}
	for _, size := range themeSizes {
		path := filepath.Join(*out, fmt.Sprintf("%s_%d.png", *name, size))
		if err := writePNG(path, iconscale.Down(master, size)); err != nil {
			return err
		}
	}
	return nil
}

// writePNG encodes picture as a PNG at path.
func writePNG(path string, picture image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(file, picture); err != nil {
		_ = file.Close()
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return file.Close()
}
