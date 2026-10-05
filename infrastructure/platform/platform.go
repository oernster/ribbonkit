// Package platform is what every ribbon's composition root does for its platform before the window
// opens, held once so each application keeps only where its icon lies. On Linux importing it sends
// GTK through X11 before Wails opens it (gtk_linux.go); Prepare hands the desktop the icon and has a
// request to end from outside end the application; GeneratingBindings says whether this is the run
// wails build makes to generate bindings (bindings_on.go).
package platform
