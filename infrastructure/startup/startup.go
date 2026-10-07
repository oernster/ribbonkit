// Package startup reads and writes the one entry that starts the application at sign-in (FR-605,
// FR-805): the Start with Windows value under HKCU on Windows, a launchd agent in the user's
// LaunchAgents folder on macOS, an XDG autostart entry on Linux.
// Settings and the Windows setup program both go through it, so the two write the same single
// entry. Each platform supplies it in a file of its own.
package startup
