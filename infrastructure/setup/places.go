//go:build windows

package setup

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"

	"github.com/oernster/ribbonkit/infrastructure/appdata"
)

// Places are the folders setup writes, resolved once (FR-810): each is under the user's own
// profile, so nothing asks for administrator rights.
type Places struct {
	// InstallDir is %LOCALAPPDATA%\Programs\<name>, where the files go (FR-802).
	InstallDir string
	// StartMenu is the user's Start Menu Programs folder, under %APPDATA%.
	StartMenu string
	// Desktop is the user's Desktop, wherever Windows has it (a Desktop moved to OneDrive included).
	Desktop string
	// Settings is the application's own settings folder, the one folder "Also forget my settings"
	// deletes (FR-806).
	Settings string
	// Exe is the application's executable in InstallDir.
	Exe string
}

const (
	// localAppData names the environment variable holding the user's local application data.
	localAppData = "LOCALAPPDATA"
	// programsSubdir is the folder under local application data that per-user programs live in.
	programsSubdir = "Programs"
)

// ErrNoLocalAppData is answered when the environment names no local application data folder.
var ErrNoLocalAppData = errors.New(localAppData + " is not set")

// KnownFolder answers the path of a Windows known folder, as windows.KnownFolderPath does.
type KnownFolder func(id *windows.KNOWNFOLDERID, flags uint32) (string, error)

// ResolvePlaces answers product's places, reading the environment through lookup and the shell's
// known folders through known.
func ResolvePlaces(product Product, lookup func(string) (string, bool), known KnownFolder) (Places, error) {
	base, ok := lookup(localAppData)
	if !ok || base == "" {
		return Places{}, ErrNoLocalAppData
	}
	settings, err := appdata.Dir(product.App, lookup)
	if err != nil {
		return Places{}, fmt.Errorf("finding the settings folder: %w", err)
	}
	startMenu, err := known(windows.FOLDERID_Programs, windows.KF_FLAG_DEFAULT)
	if err != nil {
		return Places{}, fmt.Errorf("finding the Start Menu: %w", err)
	}
	desktop, err := known(windows.FOLDERID_Desktop, windows.KF_FLAG_DEFAULT)
	if err != nil {
		return Places{}, fmt.Errorf("finding the Desktop: %w", err)
	}
	return Places{
		InstallDir: filepath.Join(base, programsSubdir, product.App.Name),
		StartMenu:  startMenu,
		Desktop:    desktop,
		Settings:   settings,
		Exe:        product.Exe(),
	}, nil
}

// Program answers the installed application's path.
func (p Places) Program() string { return filepath.Join(p.InstallDir, p.Exe) }

// Uninstaller answers the path of the setup copy left in the install folder.
func (p Places) Uninstaller() string { return filepath.Join(p.InstallDir, UninstallExeName) }
