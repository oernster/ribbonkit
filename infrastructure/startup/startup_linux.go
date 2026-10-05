package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oernster/ribbonkit/domain/identity"
)

// Ported from o7Debrief's LinuxAutostart, where both traps below were met on a real Ubuntu desktop.
//
// Inside a Flatpak, XDG_CONFIG_HOME names the app's private configuration, which the session never
// reads: an entry written there is created without error and read back as on, while nothing starts.
// So under a Flatpak the entry goes to the real ~/.config/autostart, which the manifest's
// xdg-config/autostart:create grant mounts at its true path. The session runs the entry's command
// from outside the sandbox, so for a Flatpak that command is flatpak run with the app's id.

const (
	configHomeVariable = "XDG_CONFIG_HOME"
	flatpakIDVariable  = "FLATPAK_ID"
	configFolder       = ".config"
	autostartFolder    = "autostart"
	entrySuffix        = ".desktop"
)

// The lines that mark an entry switched off. The desktop honours either over the file merely
// existing, so an entry carrying one is reported off rather than letting Settings claim otherwise.
const (
	hiddenLine        = "Hidden=true"
	gnomeDisabledLine = "X-GNOME-Autostart-enabled=false"
)

// entryTemplate is the entry written. X-GNOME-Autostart-enabled is stated so an older entry that
// switched it off cannot outlive being switched back on.
const entryTemplate = `[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Icon=%s
Terminal=false
NoDisplay=true
X-GNOME-Autostart-enabled=true
`

// New answers the entry for program, the full path of the application's executable, in the session's
// autostart folder. Under a Flatpak the command is flatpak run with its id instead.
func New(app identity.App, program string) Entry {
	dir, err := autostartDir(os.Getenv, os.UserHomeDir)
	entry := At(app, dir, program)
	if id := os.Getenv(flatpakIDVariable); id != "" {
		entry.command = "flatpak run " + id
	}
	entry.problem = err
	return entry
}

// At answers the entry for program in dir. Only a test names a folder of its own, so the real
// session's entries are never touched.
func At(app identity.App, dir, program string) Entry {
	return Entry{app: app, dir: dir, command: execQuoted(program)}
}

// entryText answers the autostart entry that runs command.
func entryText(app identity.App, command string) string {
	return fmt.Sprintf(entryTemplate, app.Name, command, app.AppID)
}

// switchedOff answers whether an entry's text carries a line the desktop reads as off.
func switchedOff(text string) bool {
	return strings.Contains(text, hiddenLine) || strings.Contains(text, gnomeDisabledLine)
}

// autostartDir answers the folder the session starts programs from. Under a Flatpak it ignores
// XDG_CONFIG_HOME, which names the sandbox's private configuration there.
func autostartDir(getenv func(string) string, home func() (string, error)) (string, error) {
	if base := getenv(configHomeVariable); base != "" && getenv(flatpakIDVariable) == "" {
		return filepath.Join(base, autostartFolder), nil
	}
	dir, err := home()
	if err != nil || dir == "" {
		return "", errNoHome
	}
	return filepath.Join(dir, configFolder, autostartFolder), nil
}

// execQuoted quotes a path for an Exec line. Inside the quotes the characters the Desktop Entry
// specification reserves are escaped with a backslash; then the value as a whole is escaped as a
// string value, which doubles every backslash.
func execQuoted(path string) string {
	quoted := `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`).Replace(path) + `"`
	return strings.ReplaceAll(quoted, `\`, `\\`)
}
