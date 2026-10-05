//go:build windows

package setup

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrNoApplication is answered when the folder named as the application holds no executable of
// the name the payload states.
var ErrNoApplication = errors.New("no application to pack")

// Payload names what the setup program carries: every file under App at the root of the install
// folder, then the licence beside them.
type Payload struct {
	// App is the folder holding the built application.
	App string
	// Exe is the application's executable, which App must hold (Product.Exe).
	Exe string
	// Licence is the licence file placed beside it as LicenceFile.
	Licence string
}

// Pack writes the payload to out as a zip archive. An application folder with no application in it
// is refused before anything is written; a file that cannot be read is refused naming it.
func Pack(out io.Writer, payload Payload) error {
	if _, err := os.Stat(filepath.Join(payload.App, payload.Exe)); err != nil {
		return fmt.Errorf("%w: %s is not in %s: %w", ErrNoApplication, payload.Exe, payload.App, err)
	}
	archive := zip.NewWriter(out)
	err := filepath.WalkDir(payload.App, func(from string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("reading %s: %w", from, err)
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(payload.App, from)
		if err != nil {
			return fmt.Errorf("placing %s: %w", from, err)
		}
		return add(archive, filepath.ToSlash(name), from)
	})
	if err != nil {
		return err
	}
	if err := add(archive, LicenceFile, payload.Licence); err != nil {
		return err
	}
	if err := archive.Close(); err != nil {
		return fmt.Errorf("finishing the payload: %w", err)
	}
	return nil
}

// add copies the file at from into the archive under name.
func add(archive *zip.Writer, name, from string) error {
	source, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	defer source.Close()
	member, err := archive.Create(name)
	if err != nil {
		return fmt.Errorf("packing %s: %w", from, err)
	}
	if _, err := io.Copy(member, source); err != nil {
		return fmt.Errorf("packing %s: %w", from, err)
	}
	return nil
}
