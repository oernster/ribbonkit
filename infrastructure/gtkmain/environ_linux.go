package gtkmain

// The environment GTK and its web view are opened in. This file holds no cgo, so it builds and its
// tests run on a machine with no GTK.

import "os"

// The backend a ribbon runs GTK on. On a Wayland session a window may not choose where it
// stands, which the ribbon must, so GTK is sent through XWayland instead (ruled 2026-09-28).
const (
	backendVariable = "GDK_BACKEND"
	backendX11      = "x11"
)

// The web view's DMABUF renderer and the value that turns it off. On NVIDIA's own driver it draws
// nothing: the window showed only its background until the renderer was off (measured on Bazzite,
// RTX 3080 Ti, 2026-10-02; turning off compositing instead did not help).
const (
	dmabufVariable = "WEBKIT_DISABLE_DMABUF_RENDERER"
	dmabufOff      = "1"
)

// ForceX11 sends GTK through X11. It must run before GTK is opened.
func ForceX11() { _ = os.Setenv(backendVariable, backendX11) }

// AvoidDMABUF turns the web view's DMABUF renderer off unless the environment already chose, so a
// user can still turn it back on. It must run before the web view starts.
func AvoidDMABUF() {
	if _, chosen := os.LookupEnv(dmabufVariable); chosen {
		return
	}
	_ = os.Setenv(dmabufVariable, dmabufOff)
}
