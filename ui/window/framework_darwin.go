package window

// Wails' macOS half uses UTType, which lives in the UniformTypeIdentifiers framework. The wails
// command adds that framework to every build it runs (read in Wails v2.12.0,
// pkg/commands/build/base.go); go build and go test do not. The window is what imports Wails, so it
// names the framework here, where no build of it can leave it out: the window's own tests and every
// application that embeds it.

// #cgo LDFLAGS: -framework UniformTypeIdentifiers
import "C"
