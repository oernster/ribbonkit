//go:build bindings

package platform

// GeneratingBindings is true in the run wails build makes to generate bindings, which carries the
// bindings build tag (read in Wails v2.12.0, pkg/commands/bindings). That run must not write the
// log, read the settings or show a tray icon.
const GeneratingBindings = true
