package appdata

import "path/filepath"

// The XDG base-directory variable for configuration, the home folder and the conventional
// configuration folder beneath home when the variable is unset. Inside a Flatpak the variable names
// the app's private configuration folder, which is where its settings belong.
const (
	configHomeVariable = "XDG_CONFIG_HOME"
	homeVariable       = "HOME"
	configFolder       = ".config"
)

// base answers the user's configuration folder; false when the environment names neither it nor a
// home folder.
func base(lookup func(string) (string, bool)) (string, bool) {
	if folder, ok := lookup(configHomeVariable); ok && folder != "" {
		return folder, true
	}
	if home, ok := lookup(homeVariable); ok && home != "" {
		return filepath.Join(home, configFolder), true
	}
	return "", false
}
