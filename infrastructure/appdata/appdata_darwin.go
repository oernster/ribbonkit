package appdata

import "path/filepath"

// The home folder and the folder beneath it macOS keeps applications' own files in.
const (
	homeVariable  = "HOME"
	supportFolder = "Library/Application Support"
)

// base answers the user's Application Support folder; false when the environment names no home.
func base(lookup func(string) (string, bool)) (string, bool) {
	home, ok := lookup(homeVariable)
	if !ok || home == "" {
		return "", false
	}
	return filepath.Join(home, supportFolder), true
}
