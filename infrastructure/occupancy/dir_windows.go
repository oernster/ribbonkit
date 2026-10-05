package occupancy

import "path/filepath"

// localVariable names the user's local application data folder: live state about windows on this
// machine, which a roaming profile must not carry to another.
const localVariable = "LOCALAPPDATA"

// dir answers %LOCALAPPDATA%\ribbonkit.
func dir(lookup func(string) (string, bool)) (string, error) {
	local, ok := lookup(localVariable)
	if !ok || local == "" {
		return "", ErrNoFolder
	}
	return filepath.Join(local, kitFolder), nil
}
