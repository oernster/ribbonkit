package occupancy

import "path/filepath"

// runtimeVariable names the user's runtime folder: private to the user and emptied at logout. It is
// the one folder two Flatpaks can both be granted (`--filesystem=xdg-run/ribbonkit:create`).
const runtimeVariable = "XDG_RUNTIME_DIR"

// dir answers $XDG_RUNTIME_DIR/ribbonkit.
func dir(lookup func(string) (string, bool)) (string, error) {
	runtime, ok := lookup(runtimeVariable)
	if !ok || runtime == "" {
		return "", ErrNoFolder
	}
	return filepath.Join(runtime, kitFolder), nil
}
