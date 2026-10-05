//go:build windows

package installer

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/oernster/ribbonkit/infrastructure/setup"
)

// page is the setup page: its markup, its scripts and its style sheet. It has no build step.
//
//go:embed page
var page embed.FS

// sheet is the page's style sheet, read here only for the two ground colours it states.
//
//go:embed page/setup.css
var sheet string

// pageRoot is the folder page holds the page under.
const pageRoot = "page"

const (
	// windowWidth and windowHeight fit the tallest screen at the sheet's type sizes (DIP). Measured
	// on 2026-09-27 by laying the page out at 804 by 661, this size less a window frame: the tallest
	// screen, Installed, took 341 of the body's 414 pixels and no screen overflowed.
	windowWidth  = 820
	windowHeight = 700
	// opaque is a colour's alpha where nothing shows through it.
	opaque = 255
)

// Window is what the application gives the setup window: its title and the pictures the page
// shows, holding every name in Pictures.
type Window struct {
	Title    string
	Pictures fs.FS
}

// Pictures are the files Window.Pictures must hold: the header mark index.html shows, then the
// theme switch's two faces, which setup-shell.js names by the appearance each switches to.
var Pictures = []string{"icon.png", "light-mode.png", "dark-mode.png"}

// Missing answers each name in Pictures that pictures does not hold; none when it holds them all.
func Missing(pictures fs.FS) []string {
	var missing []string
	for _, name := range Pictures {
		if _, err := fs.Stat(pictures, name); err != nil {
			missing = append(missing, name)
		}
	}
	return missing
}

// layered answers a file from pictures where they hold it, else from the page.
type layered struct{ pictures, page fs.FS }

// Open opens name from the pictures, falling back to the page only where the pictures lack it.
func (l layered) Open(name string) (fs.File, error) {
	file, err := l.pictures.Open(name)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		return file, err
	}
	return l.page.Open(name)
}

// Run shows the setup window over s, painted the page's own ground before the page loads so it
// never flashes the wrong one. bound is what Wails binds for the page to call: the application's
// own type embedding s, so the page reaches it as main.App.
func Run(bound any, s *Setup, window Window) error {
	light, dark, err := setup.Surfaces(sheet)
	if err != nil {
		return err
	}
	surface := light
	if s.prefersDark {
		surface = dark
	}
	assets, err := fs.Sub(page, pageRoot)
	if err != nil {
		return fmt.Errorf("reading the setup page: %w", err)
	}
	err = wails.Run(&options.App{
		Title:            window.Title,
		Width:            windowWidth,
		Height:           windowHeight,
		DisableResize:    true,
		BackgroundColour: &options.RGBA{R: surface.R, G: surface.G, B: surface.B, A: opaque},
		AssetServer:      &assetserver.Options{Assets: layered{pictures: window.Pictures, page: assets}},
		OnStartup:        s.startup,
		OnDomReady:       s.domReady,
		Bind:             []any{bound},
		Windows: &windowsoptions.Options{
			WindowClassName: s.setupID,
			// Setup's own web view cache sits under TEMP, so running it leaves no folder beside the
			// application's settings.
			WebviewUserDataPath: filepath.Join(os.TempDir(), s.setupID),
		},
	})
	if err != nil {
		return fmt.Errorf("running the setup window: %w", err)
	}
	return nil
}
