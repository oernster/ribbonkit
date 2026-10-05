// Package identity names the application ribbonkit is running for. The kit holds no product name of
// its own: each application builds one App at its composition root and hands it to the kit's
// packages that write a folder, a log, a sign-in entry or anything else a person can see by name.
package identity

// App is the application's names as the desktop knows them.
type App struct {
	// Name is the product's name as a reader sees it: the settings folder, the log, the sign-in
	// entry's name.
	Name string
	// AppID is the reverse-domain id the Linux and macOS desktops know it by: the Flatpak's id, the
	// autostart entry's file and the launchd agent's label.
	AppID string
}
