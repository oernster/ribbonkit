//go:build windows

package setup

import "github.com/oernster/ribbonkit/domain/identity"

// The endings setup gives the application's name: its executable (as wails build names it) and the
// shortcut placed in the Start Menu and on the Desktop.
const (
	exeSuffix      = ".exe"
	shortcutSuffix = ".lnk"
)

// Product is the application setup installs, as its composition root names it. Its name names the
// install folder, the executable, the shortcuts and the Apps list entry; its identity finds the
// settings folder and the Start with Windows value, as the application itself does.
type Product struct {
	App identity.App
	// Publisher is who the Apps list entry says published it.
	Publisher string
}

// Exe answers the installed application's executable.
func (p Product) Exe() string { return p.App.Name + exeSuffix }

// shortcut answers the file a shortcut to the application is saved as.
func (p Product) shortcut() string { return p.App.Name + shortcutSuffix }
